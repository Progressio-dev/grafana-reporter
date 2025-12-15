package plugin

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"
)

const (
	// htmlEmailFallbackText is the message shown to email clients that don't support HTML
	htmlEmailFallbackText = "\n\nThis email contains an embedded image. Please view it in an HTML-capable email client."
)

// EmailSender handles sending emails via SMTP
type EmailSender struct {
	host string
	port string
	user string
	pass string
	from string
}

// NewEmailSender creates a new email sender
func NewEmailSender(host, port, user, pass, from string) *EmailSender {
	return &EmailSender{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

// Send sends an email with an attachment
func (s *EmailSender) Send(to []string, subject, body string, attachment []byte, filename string) error {
	// Create message
	var buf bytes.Buffer
	
	// Write headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	
	// Create multipart writer
	writer := multipart.NewWriter(&buf)
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	buf.WriteString("\r\n")
	
	// Write body part
	bodyPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("failed to create body part: %w", err)
	}
	bodyPart.Write([]byte(body))
	
	// Write attachment part if provided
	if len(attachment) > 0 {
		// Determine content type based on filename
		contentType := "application/octet-stream"
		if strings.HasSuffix(filename, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(filename, ".pdf") {
			contentType = "application/pdf"
		}
		
		attachmentPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              []string{contentType},
			"Content-Transfer-Encoding": []string{"base64"},
			"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=%s", filename)},
		})
		if err != nil {
			return fmt.Errorf("failed to create attachment part: %w", err)
		}
		
		// Encode attachment as base64
		encoded := base64.StdEncoding.EncodeToString(attachment)
		// Write in 76-character lines
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			attachmentPart.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}
	
	writer.Close()
	
	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	
	// Send email
	if err := smtp.SendMail(addr, auth, s.from, to, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}

// SendHTML sends an email with HTML body and embedded image
func (s *EmailSender) SendHTML(to []string, subject, body string, imageData []byte, imageFormat string) error {
	// Create message
	var buf bytes.Buffer
	
	// Write headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	
	// Create outer multipart/alternative writer for text and HTML alternatives
	outerWriter := multipart.NewWriter(&buf)
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", outerWriter.Boundary()))
	buf.WriteString("\r\n")
	
	// Add plain text version for email clients that don't support HTML
	textPart, err := outerWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}
	textPart.Write([]byte(body + htmlEmailFallbackText))
	
	// Create multipart/related part for HTML with embedded images
	relatedHeader := textproto.MIMEHeader{}
	relatedWriter := multipart.NewWriter(&buf)
	relatedBoundary := relatedWriter.Boundary()
	relatedHeader.Set("Content-Type", fmt.Sprintf("multipart/related; boundary=%s", relatedBoundary))
	
	relatedPart, err := outerWriter.CreatePart(relatedHeader)
	if err != nil {
		return fmt.Errorf("failed to create related part: %w", err)
	}
	
	// Write the related part header manually
	relatedPartBuffer := &bytes.Buffer{}
	innerWriter := multipart.NewWriter(relatedPartBuffer)
	innerWriter.SetBoundary(relatedBoundary)
	
	// Write HTML body part
	htmlPart, err := innerWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/html; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("failed to create HTML body part: %w", err)
	}
	
	// Create HTML content with embedded image
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .content { margin-bottom: 20px; }
        .report-image { max-width: 100%%; height: auto; border: 1px solid #ddd; }
    </style>
</head>
<body>
    <div class="content">
        <p>%s</p>
    </div>
    <div class="report">
        <img src="cid:report-image" alt="Grafana Report" class="report-image" />
    </div>
</body>
</html>`, strings.ReplaceAll(body, "\n", "<br>"))
	
	htmlPart.Write([]byte(htmlContent))
	
	// Write embedded image part
	if len(imageData) > 0 {
		// Determine content type based on format
		contentType := "image/png"
		if imageFormat == "pdf" {
			contentType = "application/pdf"
		}
		
		imagePart, err := innerWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              []string{contentType},
			"Content-Transfer-Encoding": []string{"base64"},
			"Content-ID":                []string{"<report-image>"},
			"Content-Disposition":       []string{"inline"},
		})
		if err != nil {
			return fmt.Errorf("failed to create image part: %w", err)
		}
		
		// Encode image as base64
		encoded := base64.StdEncoding.EncodeToString(imageData)
		// Write in 76-character lines
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			imagePart.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}
	
	innerWriter.Close()
	relatedPart.Write(relatedPartBuffer.Bytes())
	
	outerWriter.Close()
	
	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	
	// Send email
	if err := smtp.SendMail(addr, auth, s.from, to, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}
