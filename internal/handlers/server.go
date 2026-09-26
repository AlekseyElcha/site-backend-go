package handlers

import (
	"log/slog"
	"site-backend-go/internal/config"
	"site-backend-go/internal/service"
)

type Server struct {
	TicketService      service.TicketService
	UserService        service.UserService
	EmailSenderService service.EmailSenderService
	FileService        service.FileService
	AddressesService   service.AddressesService
	Config             config.Config
	Logger             *slog.Logger
}

func NewServer(
	ticketService *service.TicketService,
	userService *service.UserService,
	emailSenderService service.EmailSenderService,
	fileService *service.FileService,
	addressesService *service.AddressesService,
	config *config.Config,
	logger *slog.Logger,
) *Server {
	return &Server{
		TicketService:      *ticketService,
		UserService:        *userService,
		EmailSenderService: emailSenderService,
		FileService:        *fileService,
		AddressesService:   *addressesService,
		Config:             *config,
		Logger:             logger,
	}
}
