package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/leksa/datamapper-senyar/internal/config"
)

type Service struct {
	config *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{config: cfg}
}

func (s *Service) IsConfigured() bool {
	return s.config.SMTPHost != "" && s.config.SMTPUsername != "" && s.config.SMTPPassword != ""
}

type InvitationData struct {
	RecipientName string
	InviterName   string
	OrgName       string
	Role          string
	InviteLink    string
	ExpiresIn     string
}

func (s *Service) SendInvitation(to string, data InvitationData) error {
	if !s.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	subject := fmt.Sprintf("Undangan bergabung ke %s - Dayawarga", data.OrgName)

	htmlBody, err := s.renderInvitationTemplate(data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return s.sendEmail(to, subject, htmlBody)
}

func (s *Service) renderInvitationTemplate(data InvitationData) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #1a1f2e;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background-color: #1a1f2e;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellspacing="0" cellpadding="0" style="background-color: #242a3d; border-radius: 16px; overflow: hidden;">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #4DB6AC 0%, #2B7A9E 100%); padding: 30px; text-align: center;">
                            <h1 style="margin: 0; color: #1a1f2e; font-size: 24px; font-weight: 700;">Dayawarga</h1>
                            <p style="margin: 8px 0 0 0; color: #1a1f2e; opacity: 0.8; font-size: 14px;">Platform Pemantauan Bencana</p>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #f0f1f3; font-size: 20px; font-weight: 600;">
                                Halo{{if .RecipientName}} {{.RecipientName}}{{end}},
                            </h2>
                            
                            <p style="margin: 0 0 20px 0; color: rgba(240, 241, 243, 0.8); font-size: 16px; line-height: 1.6;">
                                {{if .InviterName}}{{.InviterName}} mengundang{{else}}Anda diundang{{end}} untuk bergabung ke organisasi <strong style="color: #4DB6AC;">{{.OrgName}}</strong> sebagai <strong style="color: #4DB6AC;">{{.Role}}</strong> di platform Dayawarga.
                            </p>
                            
                            <p style="margin: 0 0 30px 0; color: rgba(240, 241, 243, 0.8); font-size: 16px; line-height: 1.6;">
                                Klik tombol di bawah untuk menerima undangan dan membuat akun Anda:
                            </p>
                            
                            <!-- CTA Button -->
                            <table role="presentation" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                <tr>
                                    <td style="background: linear-gradient(135deg, #4DB6AC 0%, #2B7A9E 100%); border-radius: 8px;">
                                        <a href="{{.InviteLink}}" target="_blank" style="display: inline-block; padding: 16px 32px; color: #1a1f2e; text-decoration: none; font-size: 16px; font-weight: 600;">
                                            Terima Undangan
                                        </a>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="margin: 30px 0 0 0; color: rgba(240, 241, 243, 0.5); font-size: 14px; line-height: 1.6;">
                                Link ini akan kedaluwarsa dalam <strong>{{.ExpiresIn}}</strong>.
                            </p>
                            
                            <p style="margin: 20px 0 0 0; color: rgba(240, 241, 243, 0.5); font-size: 14px; line-height: 1.6;">
                                Jika Anda tidak mengharapkan email ini, Anda dapat mengabaikannya.
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="padding: 20px 30px; border-top: 1px solid rgba(77, 182, 172, 0.2);">
                            <p style="margin: 0; color: rgba(240, 241, 243, 0.4); font-size: 12px; text-align: center;">
                                &copy; 2026 Dayawarga. Platform Pemantauan Bencana Berbasis Komunitas.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`

	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Service) sendEmail(to, subject, htmlBody string) error {
	from := s.config.SMTPFrom
	fromName := s.config.SMTPFromName

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	if s.config.SMTPPort == 465 {
		return s.sendEmailTLS(addr, auth, from, to, msg.String())
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg.String()))
}

func (s *Service) sendEmailTLS(addr string, auth smtp.Auth, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return client.Quit()
}
