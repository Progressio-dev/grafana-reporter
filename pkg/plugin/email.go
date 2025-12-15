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
	htmlEmailFallbackText = "\n\nThis email contains an embedded dashboard snapshot. Please view it in an HTML-capable email client."
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

// SendHTML sends an email with HTML body and embedded image (fully offline, no external links)
func (s *EmailSender) SendHTML(to []string, subject, body string, imageData []byte, imageFormat string, dashboardURL string) error {
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
	
	// Create HTML content with embedded image (fully offline, no external links)
	// Note: Email clients have security restrictions - no JavaScript, limited CSS, no iframes
	// This provides a rich, self-contained visual presentation without any external dependencies
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .email-container {
            max-width: 800px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content { 
            padding: 30px;
            line-height: 1.6;
            color: #333;
        }
        .content p {
            margin: 0 0 15px 0;
        }
        .report-section {
            margin: 30px 0;
            text-align: center;
            padding: 20px;
        }
        .report-image { 
            max-width: 100%%; 
            height: auto; 
            border: 1px solid #e0e0e0;
            border-radius: 4px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.08);
            display: block;
            margin: 0 auto;
        }
        .info-box {
            padding: 15px;
            background-color: #e8f5e9;
            border-left: 4px solid #4caf50;
            margin: 20px 30px;
            border-radius: 4px;
            font-size: 14px;
            color: #2e7d32;
        }
        .info-box strong {
            display: block;
            margin-bottom: 8px;
            font-size: 15px;
        }
        .footer {
            padding: 20px 30px;
            background-color: #f8f9fa;
            text-align: center;
            font-size: 12px;
            color: #6c757d;
            border-top: 1px solid #e9ecef;
        }
        .footer p {
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>📊 Grafana Report</h1>
        </div>
        <div class="content">
            <p>%s</p>
        </div>
        <div class="report-section">
            <img src="cid:report-image" alt="Grafana Report Dashboard" class="report-image" />
        </div>
        <div class="info-box">
            <strong>📋 Report Information</strong>
            This is a static snapshot of your Grafana dashboard, captured at the moment of generation. All content is embedded within this email for offline viewing.
        </div>
        <div class="footer">
            <p><strong>Grafana Reporter Plugin</strong></p>
            <p>This report was automatically generated and contains all data embedded within the email.</p>
        </div>
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
