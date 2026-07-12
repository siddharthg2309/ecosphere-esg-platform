package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type Sender struct{ Address, From string }

func New(address string) *Sender { return &Sender{Address: address, From: "noreply@ecosphere.local"} }
func (s *Sender) SendTemporaryPassword(ctx context.Context, name, to, password string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	subject := "Subject: Your EcoSphere account\r\n"
	headers := "From: " + s.From + "\r\nTo: " + to + "\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n"
	body := fmt.Sprintf("Hello %s,\n\nYour temporary EcoSphere password is: %s\nPlease change it after signing in.\n", name, password)
	return smtp.SendMail(s.Address, nil, s.From, []string{to}, []byte(subject+headers+body))
}
