package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"time"
)

type Service struct {
	from     string
	password string
	host     string
	port     string
	auth     smtp.Auth
}

func NewService(from, password, host, port string) *Service {
	auth := smtp.PlainAuth("", from, password, host)
	return &Service{
		from:     from,
		password: password,
		host:     host,
		port:     port,
		auth:     auth,
	}
}

type EmailData struct {
	To      string
	Subject string
	Body    string
}

// SendOTPEmail sends OTP verification email
func (s *Service) SendOTPEmail(to, fullName, otp string, expiresAt time.Time) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #6d3b2a; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .otp-box {
            font-size: 34px;
            letter-spacing: 8px;
            text-align: center;
            font-weight: bold;
            color: #6d3b2a;
            padding: 14px;
            background: white;
            border: 2px dashed #d2a26f;
            border-radius: 8px;
            margin: 18px 0;
        }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Verifikasi Akun Manscoffe</h1>
        </div>
        <div class="content">
            <p>Halo {{.Name}},</p>
            <p>Gunakan kode OTP berikut untuk verifikasi akun Anda:</p>
            <div class="otp-box">{{.OTP}}</div>
            <p>Kode ini berlaku sampai <strong>{{.Expired}}</strong>.</p>
            <p>Jika Anda tidak merasa mendaftar, abaikan email ini.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} Manscoffe. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	if fullName == "" {
		fullName = "Teman Kopi"
	}

	data := struct {
		Name    string
		OTP     string
		Expired string
		Year    int
	}{
		Name:    fullName,
		OTP:     otp,
		Expired: expiresAt.Format("02 Jan 2006 15:04"),
		Year:    time.Now().Year(),
	}

	t, err := template.New("otp").Parse(tmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	return s.send(to, "Kode OTP Verifikasi Akun", body.String())
}

// SendVerificationEmail sends email verification link
func (s *Service) SendVerificationEmail(to, token, baseURL string) error {
	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", baseURL, token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .button { 
            display: inline-block; 
            padding: 12px 30px; 
            background: #4CAF50; 
            color: white; 
            text-decoration: none; 
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Selamat Datang!</h1>
        </div>
        <div class="content">
            <h2>Verifikasi Email Anda</h2>
            <p>Terima kasih telah mendaftar di toko kami!</p>
            <p>Untuk melanjutkan, silakan verifikasi email Anda dengan mengklik tombol di bawah:</p>
            <center>
                <a href="{{.VerifyURL}}" class="button">Verifikasi Email</a>
            </center>
            <p>Atau copy link berikut ke browser Anda:</p>
            <p style="word-break: break-all; color: #666;">{{.VerifyURL}}</p>
            <p><small>Link ini akan kadaluarsa dalam 24 jam.</small></p>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas.</p>
            <p>&copy; {{.Year}} E-Commerce API. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := struct {
		VerifyURL string
		Year      int
	}{
		VerifyURL: verifyURL,
		Year:      time.Now().Year(),
	}

	t, err := template.New("verification").Parse(tmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	return s.send(to, "Verifikasi Email Anda", body.String())
}

// SendOrderConfirmation sends order confirmation email
func (s *Service) SendOrderConfirmation(to, orderID string, amount int64, items string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .order-box { 
            background: white; 
            border: 2px solid #2196F3; 
            border-radius: 8px; 
            padding: 20px; 
            margin: 20px 0; 
        }
        .amount { font-size: 24px; font-weight: bold; color: #2196F3; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Pesanan Berhasil!</h1>
        </div>
        <div class="content">
            <h2>Terima Kasih atas Pesanan Anda</h2>
            <p>Pesanan Anda telah kami terima dan sedang diproses.</p>
            
            <div class="order-box">
                <p><strong>Order ID:</strong> {{.OrderID}}</p>
                <p><strong>Total:</strong> <span class="amount">Rp {{.Amount}}</span></p>
                <p><strong>Items:</strong></p>
                <p>{{.Items}}</p>
            </div>

            <p>Kami akan mengirimkan update status pesanan Anda melalui email ini.</p>
        </div>
        <div class="footer">
            <p>Butuh bantuan? Hubungi customer service kami.</p>
            <p>&copy; {{.Year}} E-Commerce API. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := struct {
		OrderID string
		Amount  string
		Items   string
		Year    int
	}{
		OrderID: orderID,
		Amount:  formatRupiah(amount),
		Items:   items,
		Year:    time.Now().Year(),
	}

	t, err := template.New("order").Parse(tmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	return s.send(to, "Konfirmasi Pesanan - "+orderID, body.String())
}

// SendPasswordReset sends password reset email
func (s *Service) SendPasswordReset(to, token, baseURL string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #FF9800; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .button { 
            display: inline-block; 
            padding: 12px 30px; 
            background: #FF9800; 
            color: white; 
            text-decoration: none; 
            border-radius: 5px;
            margin: 20px 0;
        }
        .warning { 
            background: #fff3cd; 
            border-left: 4px solid #FF9800; 
            padding: 15px; 
            margin: 20px 0; 
        }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 Reset Password</h1>
        </div>
        <div class="content">
            <h2>Permintaan Reset Password</h2>
            <p>Kami menerima permintaan untuk mereset password akun Anda.</p>
            
            <center>
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </center>
            
            <p>Atau copy link berikut:</p>
            <p style="word-break: break-all; color: #666;">{{.ResetURL}}</p>
            
            <div class="warning">
                <strong>⚠️ Perhatian:</strong>
                <ul>
                    <li>Link ini akan kadaluarsa dalam 1 jam</li>
                    <li>Jika Anda tidak meminta reset password, abaikan email ini</li>
                    <li>Jangan bagikan link ini kepada siapapun</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>Email ini dikirim secara otomatis, mohon tidak membalas.</p>
            <p>&copy; {{.Year}} E-Commerce API. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := struct {
		ResetURL string
		Year     int
	}{
		ResetURL: resetURL,
		Year:     time.Now().Year(),
	}

	t, err := template.New("reset").Parse(tmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	return s.send(to, "Reset Password Akun Anda", body.String())
}

// send is the core email sending function
func (s *Service) send(to, subject, htmlBody string) error {
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.from, to, subject, htmlBody,
	))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(addr, s.auth, s.from, []string{to}, msg)
}

// Helper function
func formatRupiah(amount int64) string {
	// Convert cents to rupiah
	rupiah := amount / 100
	return fmt.Sprintf("%d", rupiah)
}
