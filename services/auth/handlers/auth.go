package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/strata/strata/services/auth/config"
	"github.com/strata/strata/services/auth/utils"
)

type AuthHandler struct {
	DB     *sql.DB
	RDB    *redis.Client
	Config *config.Config
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	OrgID    string `json:"org_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type sessionData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	OrgID  string `json:"org_id"`
}

func NewAuthHandler(database *sql.DB, redisClient *redis.Client, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		DB:     database,
		RDB:    redisClient,
		Config: cfg,
	}
}

// Register processes new user creation, hashing passwords with Bcrypt.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	if body.Email == "" || body.Password == "" {
		h.respondError(w, http.StatusBadRequest, "Validation Error", "Email and Password are required fields")
		return
	}

	// Assign default role and organization if missing
	role := "member"
	if body.Role != "" {
		role = body.Role
	}
	orgID := "org_default"
	if body.OrgID != "" {
		orgID = body.OrgID
	}

	// Hash password using Bcrypt cost 12
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Hashing Error", err.Error())
		return
	}
	passwordHash := string(hashedBytes)

	var id int
	var email string
	var createdRole string
	var createdOrg string
	var createdAt time.Time

	insertSQL := `
		INSERT INTO public.users (email, password_hash, role, org_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, email, role, org_id, created_at
	`
	err = h.DB.QueryRow(insertSQL, body.Email, passwordHash, role, orgID).
		Scan(&id, &email, &createdRole, &createdOrg, &createdAt)

	if err != nil {
		slog.Error("Failed to register new user", "email", body.Email, "error", err)
		h.respondError(w, http.StatusConflict, "Account Conflict", "A user account with this email address already exists")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         id,
		"email":      email,
		"role":       createdRole,
		"org_id":     createdOrg,
		"created_at": createdAt,
	})
}

// Login validates accounts and generates access + refresh tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	var (
		id           int
		email        string
		passwordHash string
		role         string
		orgID        string
		mfaEnabled   bool
	)

	query := `
		SELECT id, email, password_hash, role, org_id, mfa_enabled 
		FROM public.users 
		WHERE email = $1
	`
	err := h.DB.QueryRow(query, body.Email).
		Scan(&id, &email, &passwordHash, &role, &orgID, &mfaEnabled)

	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication Failed", "Invalid email address or password credentials")
		return
	}

	// Compare Bcrypt hashes
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)); err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication Failed", "Invalid email address or password credentials")
		return
	}

	// Create Access JWT Token and Random Refresh Token
	userIDStr := fmt.Sprintf("%d", id)
	accessToken, err := utils.GenerateAccessToken(userIDStr, email, role, orgID, h.Config.JWTSecret)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Token Error", err.Error())
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Token Error", err.Error())
		return
	}

	// Serialize and store session details in Redis with 7-day expiration
	sess := sessionData{
		UserID: userIDStr,
		Email:  email,
		Role:   role,
		OrgID:  orgID,
	}
	sessBytes, _ := json.Marshal(sess)

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", refreshToken)
	err = h.RDB.Set(ctx, sessionKey, string(sessBytes), 7*24*time.Hour).Err()
	if err != nil {
		slog.Error("Failed to store refresh token session in Redis", "error", err)
		h.respondError(w, http.StatusInternalServerError, "Session Cache Error", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600, // 1 hour
		"mfa_required":  mfaEnabled,
	})
}

// Refresh issues a new access token if the refresh token session is active.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", body.RefreshToken)
	val, err := h.RDB.Get(ctx, sessionKey).Result()
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid Session", "Refresh token is invalid or expired")
		return
	}

	var sess sessionData
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Parsing Error", err.Error())
		return
	}

	// Issue new Access Token
	newAccess, err := utils.GenerateAccessToken(sess.UserID, sess.Email, sess.Role, sess.OrgID, h.Config.JWTSecret)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Token Generation Error", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": newAccess,
		"expires_in":   3600,
	})
}

// Logout invalidates the refresh token session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", body.RefreshToken)
	h.RDB.Del(ctx, sessionKey)

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out and session terminated",
	})
}

// Me retrieves user profile metadata passed downstream by the gateway.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Gateway enriches the headers during proxy path translation
	userID := r.Header.Get("X-User-Id")
	email := r.Header.Get("X-User-Email")
	role := r.Header.Get("X-User-Role")
	orgID := r.Header.Get("X-Org-Id")

	if userID == "" {
		h.respondError(w, http.StatusUnauthorized, "Missing Identity", "Identity headers not resolved by Gateway.")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"user_id":  userID,
		"email":    email,
		"role":     role,
		"org_id":   orgID,
		"audience": "strata-auth-me",
	})
}

// MFASetup configures TOTP parameters.
func (h *AuthHandler) MFASetup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	email := r.Header.Get("X-User-Email")

	if userID == "" {
		h.respondError(w, http.StatusUnauthorized, "Missing Identity", "Authorization required to set up MFA.")
		return
	}

	// Generate 20-byte cryptographically secure random secret
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Random Key Generation Error", err.Error())
		return
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	otpauthURI := fmt.Sprintf("otpauth://totp/Strata:%s?secret=%s&issuer=Strata", email, secret)

	// Save secret to database and enable MFA flag
	updateSQL := `
		UPDATE public.users 
		SET mfa_enabled = true, mfa_secret = $1 
		WHERE id = $2
	`
	_, err := h.DB.Exec(updateSQL, secret, userID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Database Write Error", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"otpauth_url": otpauthURI,
		"message":     "MFA setup successfully enabled.",
	})
}

func (h *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) respondError(w http.ResponseWriter, status int, errType, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errType,
		"message": msg,
	})
}
