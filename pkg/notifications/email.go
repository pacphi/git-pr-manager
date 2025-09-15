package notifications

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/pacphi/git-pr-manager/pkg/merge"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// EmailNotifier sends email notifications
type EmailNotifier struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
	useTLS   bool
	logger   *utils.Logger
}

// EmailConfig contains email notification configuration
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
	UseTLS   bool
}

// EmailTemplate data for rendering email content
type EmailTemplate struct {
	Subject    string
	Results    []merge.MergeResult
	Summary    EmailSummary
	Timestamp  string
	Successful []merge.MergeResult
	Failed     []merge.MergeResult
	Skipped    []merge.MergeResult
}

// EmailSummary contains summary statistics
type EmailSummary struct {
	Total      int
	Successful int
	Failed     int
	Skipped    int
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		smtpHost: config.SMTPHost,
		smtpPort: config.SMTPPort,
		username: config.Username,
		password: config.Password,
		from:     config.From,
		to:       config.To,
		useTLS:   config.UseTLS,
		logger:   utils.GetGlobalLogger().WithComponent("email"),
	}
}

// SendMergeResults sends merge results via email
func (e *EmailNotifier) SendMergeResults(ctx context.Context, results []merge.MergeResult) error {
	if e.smtpHost == "" || len(e.to) == 0 {
		return fmt.Errorf("email configuration incomplete")
	}

	templateData := e.buildEmailTemplate(results)

	subject := fmt.Sprintf("Git PR Automation - %d PRs processed (%d merged, %d failed, %d skipped)",
		templateData.Summary.Total,
		templateData.Summary.Successful,
		templateData.Summary.Failed,
		templateData.Summary.Skipped)

	htmlBody, err := e.renderHTMLTemplate(templateData)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody := e.renderTextTemplate(templateData)

	return e.sendEmail(ctx, subject, textBody, htmlBody)
}

// SendTestMessage sends a test email to verify configuration
func (e *EmailNotifier) SendTestMessage(ctx context.Context) error {
	if e.smtpHost == "" || len(e.to) == 0 {
		return fmt.Errorf("email configuration incomplete")
	}

	subject := "Git PR Automation - Test Message"
	textBody := fmt.Sprintf(`This is a test message from Git PR Automation.

Configuration Test Results:
✅ SMTP connection successful
✅ Email delivery working

Timestamp: %s
Server: %s:%d
From: %s
To: %s

If you received this message, your email integration is configured correctly.
`, time.Now().Format("2006-01-02 15:04:05 MST"),
		e.smtpHost, e.smtpPort, e.from, strings.Join(e.to, ", "))

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Git PR Automation - Test Message</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f8f9fa; }
        .success { color: #28a745; }
        .info { background-color: #e9ecef; padding: 15px; margin: 10px 0; border-left: 4px solid #007bff; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 Git PR Automation Test</h1>
        </div>
        <div class="content">
            <p>This is a test message from Git PR Automation.</p>

            <h3>Configuration Test Results:</h3>
            <p class="success">✅ SMTP connection successful<br>
            ✅ Email delivery working</p>

            <div class="info">
                <strong>Configuration Details:</strong><br>
                <strong>Timestamp:</strong> %s<br>
                <strong>Server:</strong> %s:%d<br>
                <strong>From:</strong> %s<br>
                <strong>To:</strong> %s
            </div>

            <p>If you received this message, your email integration is configured correctly.</p>
        </div>
    </div>
</body>
</html>`, time.Now().Format("2006-01-02 15:04:05 MST"),
		e.smtpHost, e.smtpPort, e.from, strings.Join(e.to, ", "))

	return e.sendEmail(ctx, subject, textBody, htmlBody)
}

// buildEmailTemplate builds the template data for email rendering
func (e *EmailNotifier) buildEmailTemplate(results []merge.MergeResult) EmailTemplate {
	var successful, failed, skipped []merge.MergeResult

	for _, result := range results {
		if result.Error != nil {
			failed = append(failed, result)
		} else if result.Skipped {
			skipped = append(skipped, result)
		} else {
			successful = append(successful, result)
		}
	}

	return EmailTemplate{
		Results: results,
		Summary: EmailSummary{
			Total:      len(results),
			Successful: len(successful),
			Failed:     len(failed),
			Skipped:    len(skipped),
		},
		Timestamp:  time.Now().Format("2006-01-02 15:04:05 MST"),
		Successful: successful,
		Failed:     failed,
		Skipped:    skipped,
	}
}

// renderHTMLTemplate renders the HTML email template
func (e *EmailNotifier) renderHTMLTemplate(data EmailTemplate) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Git PR Automation Results</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .summary { display: flex; justify-content: space-around; background-color: #f8f9fa; padding: 20px; margin: 20px 0; }
        .stat { text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; }
        .successful { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .section { margin: 20px 0; }
        .pr-list { list-style: none; padding: 0; }
        .pr-item { background-color: #f8f9fa; margin: 10px 0; padding: 15px; border-left: 4px solid #007bff; }
        .pr-item.success { border-left-color: #28a745; }
        .pr-item.failure { border-left-color: #dc3545; }
        .pr-item.skipped { border-left-color: #ffc107; }
        .pr-title { font-weight: bold; margin-bottom: 5px; }
        .pr-meta { color: #666; font-size: 0.9em; }
        .error { color: #dc3545; font-style: italic; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 Git PR Automation Results</h1>
            <p>{{.Timestamp}}</p>
        </div>

        <div class="summary">
            <div class="stat">
                <div class="stat-number">{{.Summary.Total}}</div>
                <div>Total PRs</div>
            </div>
            <div class="stat successful">
                <div class="stat-number">{{.Summary.Successful}}</div>
                <div>Merged</div>
            </div>
            <div class="stat failed">
                <div class="stat-number">{{.Summary.Failed}}</div>
                <div>Failed</div>
            </div>
            <div class="stat skipped">
                <div class="stat-number">{{.Summary.Skipped}}</div>
                <div>Skipped</div>
            </div>
        </div>

        {{if .Successful}}
        <div class="section">
            <h2 class="successful">✅ Successfully Merged ({{len .Successful}})</h2>
            <ul class="pr-list">
                {{range .Successful}}
                <li class="pr-item success">
                    <div class="pr-title">{{.Repository}} #{{.PullRequest}}</div>
                    <div>{{.Title}}</div>
                    <div class="pr-meta">Author: {{.Author}} | Method: {{.MergeMethod}}</div>
                </li>
                {{end}}
            </ul>
        </div>
        {{end}}

        {{if .Failed}}
        <div class="section">
            <h2 class="failed">❌ Failed Merges ({{len .Failed}})</h2>
            <ul class="pr-list">
                {{range .Failed}}
                <li class="pr-item failure">
                    <div class="pr-title">{{.Repository}} #{{.PullRequest}}</div>
                    <div>{{.Title}}</div>
                    <div class="pr-meta">Author: {{.Author}}</div>
                    <div class="error">Error: {{.Error.Error}}</div>
                </li>
                {{end}}
            </ul>
        </div>
        {{end}}

        {{if .Skipped}}
        <div class="section">
            <h2 class="skipped">⏭️ Skipped PRs ({{len .Skipped}})</h2>
            <ul class="pr-list">
                {{range .Skipped}}
                <li class="pr-item skipped">
                    <div class="pr-title">{{.Repository}} #{{.PullRequest}}</div>
                    <div>{{.Title}}</div>
                    <div class="pr-meta">Author: {{.Author}}</div>
                    <div>Reason: {{.Reason}}</div>
                </li>
                {{end}}
            </ul>
        </div>
        {{end}}
    </div>
</body>
</html>`

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderTextTemplate renders the plain text email template
func (e *EmailNotifier) renderTextTemplate(data EmailTemplate) string {
	var buf strings.Builder

	buf.WriteString("Git PR Automation Results\n")
	buf.WriteString("========================\n\n")
	buf.WriteString(fmt.Sprintf("Timestamp: %s\n\n", data.Timestamp))

	buf.WriteString("Summary:\n")
	buf.WriteString(fmt.Sprintf("- Total PRs: %d\n", data.Summary.Total))
	buf.WriteString(fmt.Sprintf("- Successfully merged: %d\n", data.Summary.Successful))
	buf.WriteString(fmt.Sprintf("- Failed: %d\n", data.Summary.Failed))
	buf.WriteString(fmt.Sprintf("- Skipped: %d\n\n", data.Summary.Skipped))

	if len(data.Successful) > 0 {
		buf.WriteString("Successfully Merged PRs:\n")
		buf.WriteString("========================\n")
		for _, result := range data.Successful {
			buf.WriteString(fmt.Sprintf("✅ %s #%d: %s\n",
				result.Repository,
				result.PullRequest,
				result.Title))
			buf.WriteString(fmt.Sprintf("   Author: %s | Method: %s\n\n",
				result.Author,
				result.MergeMethod))
		}
	}

	if len(data.Failed) > 0 {
		buf.WriteString("Failed Merges:\n")
		buf.WriteString("==============\n")
		for _, result := range data.Failed {
			buf.WriteString(fmt.Sprintf("❌ %s #%d: %s\n",
				result.Repository,
				result.PullRequest,
				result.Title))
			buf.WriteString(fmt.Sprintf("   Author: %s | Error: %s\n\n",
				result.Author,
				result.Error.Error()))
		}
	}

	if len(data.Skipped) > 0 {
		buf.WriteString("Skipped PRs:\n")
		buf.WriteString("============\n")
		for _, result := range data.Skipped {
			buf.WriteString(fmt.Sprintf("⏭️ %s #%d: %s\n",
				result.Repository,
				result.PullRequest,
				result.Title))
			buf.WriteString(fmt.Sprintf("   Author: %s | Reason: %s\n\n",
				result.Author,
				result.Reason))
		}
	}

	return buf.String()
}

// sendEmail sends an email with both text and HTML content
func (e *EmailNotifier) sendEmail(_ context.Context, subject, textBody, htmlBody string) error {
	// Create message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString("--boundary123--\r\n")

	// Connect and send
	addr := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)

	err := smtp.SendMail(addr, auth, e.from, e.to, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Debug("Email notification sent successfully")
	return nil
}

// SendPRSummary sends a summary of PR check results via email
func (e *EmailNotifier) SendPRSummary(ctx context.Context, repositories []common.Repository, totalPRs, readyPRs int) error {
	if e.smtpHost == "" || len(e.to) == 0 {
		return fmt.Errorf("email configuration incomplete")
	}

	subject := fmt.Sprintf("Git PR Status Summary - %d repositories, %d PRs (%d ready)",
		len(repositories), totalPRs, readyPRs)

	textBody := fmt.Sprintf(`Git PR Automation - Status Summary
=====================================

Repository Scan Results:
- Repositories scanned: %d
- Total PRs found: %d
- PRs ready to merge: %d
- Ready percentage: %.1f%%

Timestamp: %s

This summary shows the current status of pull requests across all configured repositories.
`, len(repositories), totalPRs, readyPRs,
		float64(readyPRs)/float64(totalPRs)*100,
		time.Now().Format("2006-01-02 15:04:05 MST"))

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Git PR Status Summary</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .summary { display: flex; justify-content: space-around; background-color: #f8f9fa; padding: 20px; margin: 20px 0; }
        .stat { text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; color: #007bff; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 Git PR Status Summary</h1>
            <p>%s</p>
        </div>

        <div class="summary">
            <div class="stat">
                <div class="stat-number">%d</div>
                <div>Repositories</div>
            </div>
            <div class="stat">
                <div class="stat-number">%d</div>
                <div>Total PRs</div>
            </div>
            <div class="stat">
                <div class="stat-number">%d</div>
                <div>Ready to Merge</div>
            </div>
            <div class="stat">
                <div class="stat-number">%.1f%%</div>
                <div>Ready Rate</div>
            </div>
        </div>

        <p>This summary shows the current status of pull requests across all configured repositories.</p>
    </div>
</body>
</html>`, time.Now().Format("2006-01-02 15:04:05 MST"),
		len(repositories), totalPRs, readyPRs,
		float64(readyPRs)/float64(totalPRs)*100)

	return e.sendEmail(ctx, subject, textBody, htmlBody)
}
