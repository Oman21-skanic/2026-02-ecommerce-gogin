package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/example/ecommerce-api/internal/auth"
	"github.com/example/ecommerce-api/internal/config"
	"github.com/example/ecommerce-api/internal/email"
	"github.com/example/ecommerce-api/internal/store"
)

type AuthHandler struct {
	cfg          *config.Config
	store        store.Store
	jwtManager   *auth.JWTManager
	emailService *email.Service
}

func NewAuthHandler(cfg *config.Config, st store.Store, jm *auth.JWTManager, es *email.Service) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		store:        st,
		jwtManager:   jm,
		emailService: es,
	}
}

type signupReq struct {
	FullName        string `json:"full_name" binding:"required,min=2"`
	Phone           string `json:"phone" binding:"required,min=8"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data signup tidak valid"})
		return
	}

	if req.ConfirmPassword != "" && req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konfirmasi password tidak cocok"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}

	u, err := h.store.CreateUser(req.FullName, req.Phone, req.Email, hash, "user", "email", "", false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otp := generateOTP()
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := h.store.CreateEmailVerification(u.ID, otp, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat OTP verifikasi"})
		return
	}

	if h.emailService != nil {
		go func() {
			_ = h.emailService.SendOTPEmail(req.Email, req.FullName, otp, expiresAt)
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Akun berhasil dibuat. Silakan verifikasi OTP dari email.",
		"user_id":       u.ID,
		"otp_expiresAt": expiresAt,
	})
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	u, err := h.store.GetUserByEmail(req.Email)
	if err != nil || !auth.VerifyPassword(u.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	if !u.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Email belum terverifikasi. Silakan verifikasi OTP terlebih dahulu."})
		return
	}

	isAdmin := u.Role == "admin"
	t, err := h.jwtManager.Generate(u.ID, u.Email, isAdmin, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":          t,
		"full_name":      u.FullName,
		"phone":          u.Phone,
		"email":          u.Email,
		"role":           u.Role,
		"email_verified": u.EmailVerified,
	})
}

type verifyOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,min=4"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan OTP wajib diisi"})
		return
	}

	verification, err := h.store.GetEmailVerification(strings.TrimSpace(req.OTP))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP tidak valid"})
		return
	}

	user, err := h.store.GetUserByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User tidak ditemukan"})
		return
	}

	if verification.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP tidak sesuai dengan email"})
		return
	}

	if time.Now().After(verification.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP sudah kadaluarsa"})
		return
	}

	if err := h.store.MarkEmailVerified(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal verifikasi email"})
		return
	}

	_ = h.store.DeleteEmailVerification(verification.Token)

	c.JSON(http.StatusOK, gin.H{"message": "Email berhasil diverifikasi"})
}

type resendOTPReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req resendOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email wajib diisi"})
		return
	}

	user, err := h.store.GetUserByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, OTP baru akan dikirim"})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"message": "Email sudah terverifikasi"})
		return
	}

	otp := generateOTP()
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := h.store.CreateEmailVerification(user.ID, otp, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat OTP baru"})
		return
	}

	if h.emailService != nil {
		go func() {
			_ = h.emailService.SendOTPEmail(user.Email, user.FullName, otp, expiresAt)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP baru telah dikirim"})
}

type googleLoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name"`
	GoogleID string `json:"google_id"`
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req googleLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload Google login tidak valid"})
		return
	}

	user, err := h.store.GetUserByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		tempPasswordHash, hashErr := auth.HashPassword(uuid.NewString())
		if hashErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses akun Google"})
			return
		}

		fullName := strings.TrimSpace(req.FullName)
		if fullName == "" {
			fullName = "Google User"
		}

		user, err = h.store.CreateUser(fullName, "", strings.TrimSpace(req.Email), tempPasswordHash, "user", "google", strings.TrimSpace(req.GoogleID), true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	isAdmin := user.Role == "admin"
	token, err := h.jwtManager.Generate(user.ID, user.Email, isAdmin, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":          token,
		"email":          user.Email,
		"full_name":      user.FullName,
		"role":           user.Role,
		"email_verified": true,
	})
}

func (h *AuthHandler) GoogleStart(c *gin.Context) {
	conf := h.googleOAuthConfig()
	if conf == nil {
		h.redirectToFrontend(c, "/auth/login", "Google OAuth belum dikonfigurasi", "")
		return
	}

	next := sanitizeNext(c.Query("next"))
	state := encodeState(next)
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	conf := h.googleOAuthConfig()
	if conf == nil {
		h.redirectToFrontend(c, "/auth/login", "Google OAuth belum dikonfigurasi", "")
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	next := decodeState(state)
	if code == "" {
		h.redirectToFrontend(c, next, "Google login gagal: code kosong", "")
		return
	}

	ctx := context.Background()
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Printf("google oauth exchange error: %v", err)
		h.redirectToFrontend(c, next, "Google login gagal: token exchange", "")
		return
	}

	client := conf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json")
	if err != nil {
		log.Printf("google oauth userinfo error: %v", err)
		h.redirectToFrontend(c, next, "Gagal mengambil profil Google", "")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("google oauth userinfo status: %d", resp.StatusCode)
		h.redirectToFrontend(c, next, "Profil Google tidak valid", "")
		return
	}

	var payload struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		h.redirectToFrontend(c, next, "Gagal membaca profil Google", "")
		return
	}

	if !payload.VerifiedEmail || payload.Email == "" {
		h.redirectToFrontend(c, next, "Email Google belum terverifikasi", "")
		return
	}

	user, err := h.store.GetUserByEmail(strings.TrimSpace(payload.Email))
	if err != nil {
		tempPasswordHash, hashErr := auth.HashPassword(uuid.NewString())
		if hashErr != nil {
			h.redirectToFrontend(c, next, "Gagal memproses akun Google", "")
			return
		}
		fullName := strings.TrimSpace(payload.Name)
		if fullName == "" {
			fullName = "Google User"
		}
		user, err = h.store.CreateUser(fullName, "", payload.Email, tempPasswordHash, "user", "google", payload.ID, true)
		if err != nil {
			h.redirectToFrontend(c, next, err.Error(), "")
			return
		}
	}

	isAdmin := user.Role == "admin"
	jwtToken, err := h.jwtManager.Generate(user.ID, user.Email, isAdmin, 24*time.Hour)
	if err != nil {
		h.redirectToFrontend(c, next, "Token error", "")
		return
	}

	params := url.Values{}
	params.Set("token", jwtToken)
	params.Set("email", user.Email)
	params.Set("role", user.Role)
	params.Set("next", next)
	frontend := h.cfg.FrontendURL
	if frontend == "" {
		frontend = "http://localhost:4321"
	}
	callbackURL := fmt.Sprintf("%s/api/auth/google-callback?%s", strings.TrimRight(frontend, "/"), params.Encode())
	if payload.Picture != "" {
		callbackURL = callbackURL + "&photo=" + url.QueryEscape(payload.Picture)
	}

	c.Redirect(http.StatusFound, callbackURL)
}

func (h *AuthHandler) googleOAuthConfig() *oauth2.Config {
	if h.cfg.GoogleClientID == "" || h.cfg.GoogleClientSecret == "" || h.cfg.GoogleRedirectURL == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  h.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (h *AuthHandler) redirectToFrontend(c *gin.Context, next, errMsg, okMsg string) {
	frontend := h.cfg.FrontendURL
	if frontend == "" {
		frontend = "http://localhost:4321"
	}
	params := url.Values{}
	if errMsg != "" {
		params.Set("error", errMsg)
	}
	if okMsg != "" {
		params.Set("message", okMsg)
	}
	if next != "" {
		params.Set("next", next)
	}
	redirectURL := fmt.Sprintf("%s/auth/login?%s", strings.TrimRight(frontend, "/"), params.Encode())
	c.Redirect(http.StatusFound, redirectURL)
}

func sanitizeNext(next string) string {
	if next == "" {
		return "/products"
	}
	if !strings.HasPrefix(next, "/") {
		return "/products"
	}
	return next
}

func encodeState(next string) string {
	state := fmt.Sprintf("%s|%s", uuid.NewString(), sanitizeNext(next))
	return base64.RawURLEncoding.EncodeToString([]byte(state))
}

func decodeState(state string) string {
	if state == "" {
		return "/products"
	}
	decoded, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "/products"
	}
	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return "/products"
	}
	return sanitizeNext(parts[1])
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	verification, err := h.store.GetEmailVerification(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token tidak valid atau sudah kadaluarsa"})
		return
	}

	if time.Now().After(verification.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token sudah kadaluarsa"})
		return
	}

	// Mark email as verified
	if err := h.store.MarkEmailVerified(verification.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Delete verification token
	_ = h.store.DeleteEmailVerification(token)

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Email berhasil diverifikasi! Silakan login.",
	})
}

type requestResetReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req requestResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email harus valid"})
		return
	}

	u, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, link reset password akan dikirim"})
		return
	}

	// Create reset token
	token := uuid.NewString()
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := h.store.CreatePasswordReset(u.ID, token, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	// Send reset email
	if h.emailService != nil {
		go func() {
			_ = h.emailService.SendPasswordReset(req.Email, token, h.cfg.BaseURL)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Link reset password telah dikirim ke email Anda"})
}

type resetPasswordReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token dan password baru diperlukan (min 6 karakter)"})
		return
	}

	reset, err := h.store.GetPasswordReset(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token tidak valid"})
		return
	}

	if time.Now().After(reset.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token sudah kadaluarsa"})
		return
	}

	// Hash new password
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.store.UpdateUserPassword(reset.UserID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Delete reset token
	_ = h.store.DeletePasswordReset(req.Token)

	c.JSON(http.StatusOK, gin.H{"message": "✅ Password berhasil direset! Silakan login dengan password baru."})
}

func generateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}
