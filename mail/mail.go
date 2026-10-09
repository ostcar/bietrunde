package mail

import (
	"fmt"

	"github.com/ostcar/bietrunde/config"
	"github.com/wneessen/go-mail"
)

type Msg = mail.Msg

func CreateMail(cfg config.Config, to string, ccVorstand bool, subject string, format string, a ...any) (*mail.Msg, error) {
	text := fmt.Sprintf(format, a...)

	m := mail.NewMsg()
	if err := m.From(cfg.MailFrom); err != nil {
		return nil, fmt.Errorf("set from header: %w", err)
	}

	if err := m.To(to); err != nil {
		return nil, fmt.Errorf("set to header: %w", err)
	}

	if ccVorstand {
		if err := m.Cc(cfg.MailVorstand); err != nil {
			return nil, fmt.Errorf("set to-cc header: %w", err)
		}
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, text)

	return m, nil
}

func SendMails(debug bool, mails ...*mail.Msg) error {
	if debug {
		return sendMailsDebug(mails...)
	}

	c, err := mail.NewClient("localhost",
		mail.WithPort(25),
		mail.WithSMTPAuth(mail.SMTPAuthNoAuth),
		mail.WithTLSPolicy(mail.NoTLS),
	)
	if err != nil {
		return fmt.Errorf("create mail client: %w", err)
	}

	if err := c.DialAndSend(mails...); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

func sendMailsDebug(mails ...*mail.Msg) error {
	for _, m := range mails {
		fmt.Printf("To: %v\n", m.GetTo())
		fmt.Printf("Cc: %v\n", m.GetCc())

		fmt.Printf("Subject: %s\n", m.GetGenHeader(mail.HeaderSubject))
		body, err := m.GetParts()[0].GetContent()
		if err != nil {
			return fmt.Errorf("create body: %w", err)
		}
		fmt.Printf("Body:\n%s\n\n", body)
	}

	return nil
}
