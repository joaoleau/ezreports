package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/joaoleau/ezreports/internal/config"
)

type SMTPService struct {
	SMTPConfig config.SMTPConfig
}

func (s *SMTPService) SendMail(
	to []string,
	subject string,
	htmlBody string,
	attachments []string,
) error {

	var buf bytes.Buffer
	boundary := "PROMGO_BOUNDARY"

	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.SMTPConfig.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	buf.WriteString(htmlBody + "\r\n")

	for _, file := range attachments {
		if err := s.attachFile(&buf, boundary, file); err != nil {
			return err
		}
	}

	buf.WriteString(fmt.Sprintf("--%s--", boundary))

	auth := smtp.PlainAuth(
		"",
		s.SMTPConfig.Username,
		s.SMTPConfig.Password,
		s.SMTPConfig.Host,
	)

	addr := fmt.Sprintf("%s:%d", s.SMTPConfig.Host, s.SMTPConfig.Port)
	return smtp.SendMail(addr, auth, s.SMTPConfig.From, to, buf.Bytes())
}

func (s *SMTPService) attachFile(
	buf *bytes.Buffer,
	boundary string,
	filePath string,
) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	filename := filepath.Base(filePath)
	mimeType := "application/octet-stream"

	if strings.HasSuffix(filename, ".pdf") {
		mimeType = "application/pdf"
	}
	if strings.HasSuffix(filename, ".png") {
		mimeType = "image/png"
	}

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", mimeType, filename))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", filename))

	encoded := base64.StdEncoding.EncodeToString(data)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[i:end] + "\r\n")
	}

	return nil
}
