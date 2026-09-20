package main

import (
	"fmt"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/config"
	"site-backend-go/internal/db"
	"site-backend-go/internal/handlers"
	"site-backend-go/internal/redis_client"
	"site-backend-go/internal/service"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// dbConnString := "postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable"
	dbPgConnString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	//
	//ctx := context.Background()

	sqlDB, err := db.InitDB(dbPgConnString)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	isLocalMode := true
	resendKey := cfg.Email.ResendAPIKey
	fromEmail := cfg.Email.FromEmail

	uploadDir := cfg.Files.UploadDir

	dbStorage := db.NewStorage(sqlDB)
	rawRedisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	rdb := redis_client.NewRedisClient(rawRedisClient)

	ticketService := service.NewTicketService(dbStorage)
	userService := service.NewUserService(dbStorage)
	emailSenderService := service.NewEmailSenderService(isLocalMode, resendKey, fromEmail)
	filesService := service.NewFileService(uploadDir, dbStorage)
	authHandlers := auth.NewAuthHandler(rdb, userService)

	appServer := handlers.NewServer(
		ticketService,
		userService,
		emailSenderService,
		filesService,
		cfg,
	)

	mux := http.NewServeMux()
	fmt.Println("Starting server...")

	// Получение всех тикетов
	mux.HandleFunc("/api/v1/tickets/get", appServer.GetAllTicketsInfoHandler)
	// Получение тикета по его ID
	mux.HandleFunc("GET /api/v1/tickets/get/{id}", appServer.GetTicketInfoByIDHandler)
	// Создание нового тикета
	mux.HandleFunc("POST /api/v1/tickets/create", appServer.CreateTicketHandler)
	// Получение всех обращений пользователя по ID

	//
	mux.HandleFunc("POST /api/v1/tickets/answer", appServer.AnswerTicketByIDHandler)
	//
	mux.HandleFunc("PUT /api/v1/tickets/status/update", appServer.UpdateTicketStatusByIDHandler)

	//
	mux.HandleFunc("GET /api/v1/extra_messages/get", appServer.GetExtraMessagesForTicketByID)
	//
	mux.HandleFunc("POST /api/v1/extra_messages/create", appServer.CreateExtraMessageHandler)
	//
	mux.HandleFunc("GET /api/v1/users/get/{id}", appServer.GetUserInfoByIDHandler)
	//
	mux.HandleFunc("POST /api/v1/users/create", appServer.CreateNewUserHandler)
	//
	mux.HandleFunc("PUT /api/v1/users/update", appServer.UpdateUserInfoByIDHandler)

	//
	mux.HandleFunc("GET /api/v1/auth/info", authHandlers.GetSelfInfoFromCookies)
	//
	mux.HandleFunc("POST /api/v1/auth/code", authHandlers.RequestAuthCodeHandler)
	//
	mux.HandleFunc("POST /api/v1/auth/login", authHandlers.LoginHandler)
	//
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandlers.RefreshHandler)
	//
	mux.HandleFunc("POST /api/v1/auth/logout", authHandlers.LogoutHandler)

	//
	mux.HandleFunc("POST /api/v1/files/upload", appServer.UploadFileHandler)
	//
	mux.HandleFunc("GET /api/v1/files/download", appServer.DownloadFileByIDHandler)

	//
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	srv := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.App.Port),
		Handler: mux,
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
