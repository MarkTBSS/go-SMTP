package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

// loginAuth is an implementation of smtp.Auth for the LOGIN authentication mechanism.
type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
		}
	}
	return nil, nil
}

// sendEmail sends an email using the provided SMTP server credentials with LOGIN authentication.
func sendEmail(to, subject, body, from, password, smtpHost, smtpPort string) error {
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject

	// Combine headers and body into a single message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Set up authentication information
	auth := &loginAuth{username: from, password: password}

	// Connect to the SMTP server
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Close()

	// Send EHLO command
	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("failed to send EHLO command: %v", err)
	}

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpHost,
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %v", err)
		}
	}

	// Authenticate using LOGIN
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Send the email
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data connection: %v", err)
	}

	return nil
}

func main() {
	to := "aiya.ai@astartechs.com"
	subject := "Subject of the Email"
	body := "This is the body of the email."
	from := "dpomate@pea.co.th"
	password := "PEA@dpo@2024"
	smtpHost := "email.pea.co.th"
	smtpPort := "587"

	err := sendEmail(to, subject, body, from, password, smtpHost, smtpPort)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		fmt.Println("Email sent successfully!")
	}
}
