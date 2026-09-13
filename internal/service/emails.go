package service

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"
)

type EmailSenderService interface {
	Send(ctx context.Context, to []string, subject, html string) (string, error)
}

func NewEmailSenderService(isLocal bool, apiKey, from string) EmailSenderService {
	if isLocal {
		return NewMockSender()
	}
	return NewResendSender(apiKey, from)
}

type ResendEmailSender struct {
	client *resend.Client
	from   string
}

func NewResendSender(apiKey, from string) *ResendEmailSender {
	return &ResendEmailSender{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

func (s *ResendEmailSender) Send(ctx context.Context, to []string, subject, html string) (string, error) {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      to,
		Subject: subject,
		Html:    html,
	}
	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", err
	}
	return sent.Id, nil
}

type MockEmailSender struct{}

func NewMockSender() *MockEmailSender {
	return &MockEmailSender{}
}

func (s *MockEmailSender) Send(ctx context.Context, to []string, subject, html string) (string, error) {
	fmt.Println("\n--- [LOCAL EMAIL LOG] ---")
	fmt.Printf("Кому: %v\nТема: %s\nТело: %s\n", to, subject, html)
	fmt.Println("-------------------------\n")
	return "mock-id-12345", nil
}
