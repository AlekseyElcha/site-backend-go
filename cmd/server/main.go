package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/config"
	"site-backend-go/internal/db"
	"site-backend-go/internal/handlers"
	"site-backend-go/internal/middleware"
	"site-backend-go/internal/redis_client"
	"site-backend-go/internal/service"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var Version = "1.0.0"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logLevelCfg := cfg.Logger.Level
	var logLevel slog.Level

	err = logLevel.UnmarshalText([]byte(strings.ToLower(logLevelCfg)))
	if err != nil {
		logLevel = slog.LevelInfo

		fmt.Println("error parsing log level, using LevelInfo as default: " + err.Error())
	}

	loggerOpts := &slog.HandlerOptions{
		Level: logLevel,
		// AddSource: true,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))

	// postgres db
	dbPgConnString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	sqlDB, err := db.InitDB(dbPgConnString)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	publicKeyBytes, err := os.ReadFile("certs/public.pem")
	if err != nil {
		panic(fmt.Errorf("failed to read public key: %w", err))
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		panic(fmt.Errorf("public key error %v", err))
	}

	isLocalMode := cfg.Mode.IsLocal

	resendKey := cfg.Email.ResendAPIKey
	fromEmail := cfg.Email.FromEmail

	uploadDir := cfg.Files.UploadDir

	dbStorage := db.NewStorage(sqlDB, logger)
	rawRedisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	rdb := redis_client.NewRedisClient(rawRedisClient)

	ticketService := service.NewTicketService(dbStorage, logger)
	userService := service.NewUserService(dbStorage, logger)
	emailSenderService := service.NewEmailSenderService(isLocalMode, resendKey, fromEmail)
	filesService := service.NewFileService(uploadDir, dbStorage)
	addressesService := service.NewAddressesService(dbStorage)
	authHandlers := auth.NewAuthHandler(cfg, rdb, userService)

	appServer := handlers.NewServer(
		ticketService,
		userService,
		emailSenderService,
		filesService,
		addressesService,
		cfg,
		logger,
	)

	mux := http.NewServeMux()
	logger.Info("Starting server", "port", cfg.App.Port)

	mm := middleware.NewMiddlewareManager(
		logger,
		verifyKey,
	)

	// Получение всех тикетов
	mux.Handle("GET /api/v1/tickets/get", mm.AuthCheckMiddleware([]string{"admin"})(
		http.HandlerFunc(appServer.GetAllTicketsInfoHandler),
	))

	// Получение тикета по его ID
	mux.Handle("GET /api/v1/tickets/get/{id}", mm.AuthCheckMiddleware([]string{"admin", "user"})(
		http.HandlerFunc(appServer.GetTicketInfoByIDHandler),
	))

	// Создание нового тикета
	mux.Handle("POST /api/v1/tickets/create", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.CreateTicketHandler),
	))

	// Получение всех обращений пользователя по ID
	mux.Handle("GET /api/v1/tickets/user/{user_id}", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.GetTicketsByUserIDHandler),
	))

	// Ответ на обращение пользователя
	mux.Handle("POST /api/v1/tickets/answer", mm.AuthCheckMiddleware([]string{"admin"})(
		http.HandlerFunc(appServer.AnswerTicketByIDHandler),
	))

	// Обновление статуса обращения по его ID
	mux.Handle("PUT /api/v1/tickets/status/update", mm.AuthCheckMiddleware([]string{"admin"})(
		http.HandlerFunc(appServer.UpdateTicketStatusByIDHandler),
	))

	// Получение всех дполнительных сообщений пользователя по ID обращения
	//mux.Handle("POST /api/v1/extra_messages/get", mm.AuthCheckMiddleware([]string{"user", "admin"})(
	//	http.HandlerFunc(appServer.GetExtraMessagesForTicketByID),
	//))

	// Создание дополнительного сообщения
	mux.Handle("POST /api/v1/extra_messages/create", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.CreateExtraMessageHandler),
	))

	// Получение информации о пользователе по его ID
	mux.Handle("GET /api/v1/users/get/{id}", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.GetUserInfoByIDHandler),
	))

	// Создание нового пользователя
	mux.Handle("POST /api/v1/users/create", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.CreateNewUserHandler),
	))

	// Обноление информации о пользователе по его ID
	mux.Handle("PUT /api/v1/users/update", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.UpdateUserInfoByIDHandler),
	))

	// Информация по авторизации
	mux.HandleFunc("GET /api/v1/auth/info", authHandlers.GetSelfInfoFromCookies)

	// Запрос кода на email для аутентификации
	mux.HandleFunc("POST /api/v1/auth/code", authHandlers.RequestAuthCodeHandler)

	// Логин, тут без комментариев
	mux.HandleFunc("POST /api/v1/auth/login", authHandlers.LoginHandler)

	// Продление сессии пользователя
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandlers.RefreshHandler)

	// Логаут, буквально
	mux.HandleFunc("POST /api/v1/auth/logout", authHandlers.LogoutHandler)

	// Загрузка файла на сервер
	mux.Handle("POST /api/v1/files/upload", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.UploadFileHandler),
	))

	// Скачивание файла с сервера
	mux.Handle("POST /api/v1/files/download", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.DownloadFileByIDHandler),
	))

	// Поиск улиц среди доступных адресов
	mux.HandleFunc("GET /api/v1/address/find/street", appServer.FindSuitableStreetsHandler)

	// Поиск домов среди доступных адресов
	mux.HandleFunc("GET /api/v1/address/find/house", appServer.FindSuitableHousesHandler)

	// Полный поиск по улица + дом + подъезд
	mux.HandleFunc("GET /api/v1/address/find/complete", appServer.FindSuitableAddressesHandler)

	// Поиск по обращениям через ILIKE PostgreSQL, передача строки поиска через Query-parameter: search
	mux.Handle("POST /api/v1/tickets/search", mm.AuthCheckMiddleware([]string{"user", "admin"})(
		http.HandlerFunc(appServer.SearchTicketsAlikeHandler),
	))

	// Раздача фронта
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	wrappedMux := mm.LoggingAndRequestIDMiddleware(mux)

	srv := http.Server{
		// Addr:    ":" + strconv.Itoa(cfg.App.Port),
		Addr:    "0.0.0.0:" + strconv.Itoa(cfg.App.Port),
		Handler: wrappedMux,
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
