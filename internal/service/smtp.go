package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"crypto/tls"
	"github.com/joaoleau/ezreports/internal/config"
)

type SMTPService struct {
	SMTPConfig config.SMTPConfig
}

func (s *SMTPService) writeAttachment(buf *bytes.Buffer, boundary, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileName := filepath.Base(filePath)

	var contentType string
	switch {
	case strings.HasSuffix(fileName, ".png"):
		contentType = "image/png"
	case strings.HasSuffix(fileName, ".pdf"):
		contentType = "application/pdf"
	default:
		return fmt.Errorf("tipo nao suportado: %s", fileName)
	}

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, fileName))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", fileName))

	buf.Write(encoded)
	buf.WriteString("\r\n")

	return nil
}

func (s *SMTPService) buildMessage(from string, to []string, subject string, filenames []string, text string) []byte {
	var buffer bytes.Buffer

	boundary := "EZREPORT_BOUNDARY"

	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

	// Corpo do email
	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	buffer.WriteString(text + "\r\n")

	// Anexos
	for _, file := range filenames {
		_ = s.writeAttachment(&buffer, boundary, file)
	}

	buffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buffer.Bytes()
}

func (s *SMTPService) sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	return w.Close()
}

func (s *SMTPService) SendSimpleEmail(to []string, subject string, body string,	attachments []string,) error {
	auth := smtp.PlainAuth(
		"",
		s.SMTPConfig.Username,
		s.SMTPConfig.Password,
		s.SMTPConfig.Host,
	)

	addr := fmt.Sprintf("%s:%d", s.SMTPConfig.Host, s.SMTPConfig.Port)
	
	msg := s.buildMessage(s.SMTPConfig.From, to, subject, attachments, body)

	// Porta 465 (SMTPS) SES
	if s.SMTPConfig.Port == 465 {
		return s.sendMailTLS(addr, auth, s.SMTPConfig.From, to, msg, s.SMTPConfig.Host)
	}

	return smtp.SendMail(addr, auth, s.SMTPConfig.From, to, msg)
}