package main

import (
	"fmt"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/db"
	"site-backend-go/internal/handlers"
	"site-backend-go/internal/service"
)

func main() {
	dbConnString := "postgres://postgres:postgres@localhost:5432/tickets?sslmode=disable"

	sqlDB, err := db.InitDB(dbConnString)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	isLocalMode := true
	resendKey := "re_xxxxxxxxx"
	fromEmail := "Acme <info@domain.com>"

	uploadDir := "./files"

	dbStorage := db.NewStorage(sqlDB)

	ticketService := service.NewTicketService(dbStorage)
	userService := service.NewUserService(dbStorage)
	emailSenderService := service.NewEmailSenderService(isLocalMode, resendKey, fromEmail)
	filesService := service.NewFileService(uploadDir)

	appServer := handlers.NewServer(
		*ticketService,
		*userService,
		emailSenderService,
		*filesService,
	)

	mux := http.NewServeMux()
	fmt.Println("Starting server...")

	// mux.Handle("POST /create", auth.AuthCheckMiddleware("user", http.HandlerFunc(appServer.CreateTicketHandler)))
	mux.HandleFunc("GET /api/v1/tickets/get/{id}", appServer.GetTicketInfoByIDHandler)
	mux.HandleFunc("POST /api/v1/tickets/create", appServer.CreateTicketHandler)
	mux.HandleFunc("PUT /api/v1/tickets/update", appServer.UpdateTicketStatusByIDHandler)
	mux.HandleFunc("POST /api/v1/tickets/answer", appServer.AnswerTicketByIDHandler)

	mux.HandleFunc("POST /api/v1/users/create", appServer.CreateNewUserHandler)

	mux.HandleFunc("POST /api/v1/auth/login", auth.LoginHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", auth.LoginHandler)

	mux.HandleFunc("POST /api/v1/files/upload", auth.LoginHandler)

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()

}
