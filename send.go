package email

import (
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"sync"
)

// MailBox type
type MailBox struct {
	envelopes chan Email
	closed    chan struct{}
	closer    sync.Once
}

// NewMailBox returns a mailbox with own queue
func NewMailBox() *MailBox {
	mailBox := &MailBox{
		envelopes: make(chan Email),
		closed:    make(chan struct{}),
	}

	go func() {
		for {
			select {
			case envelope := <-mailBox.envelopes:
				_ = Send(&envelope)
			case <-mailBox.closed:
				return
			}
		}
	}()

	return mailBox
}

// Close mailbox, use closer to avoid possible race condition.
// Throws away all messages left in pipe
func (m *MailBox) Close() {
	m.closer.Do(func() {
		close(m.closed)
	})
}

// Put new email in mail queue
func (m *MailBox) Put(e Email) error {
	select {
	case m.envelopes <- e:
		return nil
		// sent successfully
	case <-m.closed:
		return errors.New("Put on closed mailbox")
	}
}

// Send sends email using the values specified in config
func Send(e *Email) error {
	if config.Server == "" {
		return errors.New("No smtp server")
	}

	addr := fmt.Sprintf("%s:%d", config.Server, config.Port)
	if config.SMTPPassword != "" && config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.Server)
		return send(e, addr, auth)
	}
	return SendSMTPWithoutAuth(e, addr)
}

// SendSMTP sends email to specified smtp address using given auth
func SendSMTP(e *Email, addr string, auth smtp.Auth) error {
	if addr == "" {
		return errors.New("No smtp server")
	}

	return send(e, addr, auth)
}

// SendSMTPWithoutAuth sends email to specified smtpServer without trying to authenticate
func SendSMTPWithoutAuth(e *Email, smtpServer string) error {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)

	from, err := mail.ParseAddress(e.From)
	if err != nil {
		return err
	}

	raw, err := e.Bytes()
	if err != nil {
		return err
	}

	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("Must specify at least one From address and one To address")
	}

	// Create the smtp connection
	smtpClient, err := smtp.Dial(smtpServer)
	if err != nil {
		return err
	}
	defer smtpClient.Quit()

	// To && From
	if err = smtpClient.Mail(from.Address); err != nil {
		return err
	}

	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		if err = smtpClient.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	smtpData, err := smtpClient.Data()
	if err != nil {
		return err
	}

	if _, err = smtpData.Write(raw); err != nil {
		return err
	}

	err = smtpData.Close()
	return err
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
