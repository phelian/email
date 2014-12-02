package email

import (
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
)

// Send sends email using the values specified in config
func Send(e *Email) error {
	addr := fmt.Sprintf("%s:%d", config.Server, config.Port)
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.Server)
	return send(e, addr, auth)
}

// SendTo sends email to specified address using given auth
func SendSmtpAddr(e *Email, addr string, auth smtp.Auth) error {
	return send(e, addr, auth)
}

// Send an email using the given host and SMTP auth (optional), returns any error thrown by smtp.SendMail
// This function merges the To, Cc, and Bcc fields and calls the smtp.SendMail function using the Email.Bytes() output as the message
func send(e *Email, addr string, a smtp.Auth) error {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("Must specify at least one From address and one To address")
	}
	from, err := mail.ParseAddress(e.From)
	if err != nil {
		return err
	}
	raw, err := e.Bytes()
	if err != nil {
		return err
	}
	return smtp.SendMail(addr, a, from.Address, to, raw)
}
