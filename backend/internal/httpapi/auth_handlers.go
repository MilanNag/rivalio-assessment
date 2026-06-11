package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/milann/taskflow/internal/auth"
	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/store"
)

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields := validateSignup(req.Email, req.Password); fields != nil {
		writeValidationError(w, fields)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("hash password", "error", err)
		writeInternalError(w)
		return
	}

	// Bootstrap convenience: the very first registered user becomes admin.
	role := models.RoleUser
	if count, err := s.store.CountUsers(r.Context()); err == nil && count == 0 {
		role = models.RoleAdmin
	}

	user, err := s.store.CreateUser(r.Context(), req.Email, hash, role)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateEmail) {
			writeError(w, http.StatusConflict, "email_taken", "An account with this email already exists.")
			return
		}
		s.logger.Error("create user", "error", err)
		writeInternalError(w)
		return
	}

	token, err := auth.GenerateToken(s.cfg.JWTSecret, user.ID, user.Email, user.Role, s.cfg.JWTExpiry)
	if err != nil {
		s.logger.Error("generate token", "error", err)
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, User: user})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {
		writeValidationError(w, map[string]string{
			"email":    "Email is required.",
			"password": "Password is required.",
		})
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
			return
		}
		s.logger.Error("get user", "error", err)
		writeInternalError(w)
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
		return
	}

	token, err := auth.GenerateToken(s.cfg.JWTSecret, user.ID, user.Email, user.Role, s.cfg.JWTExpiry)
	if err != nil {
		s.logger.Error("generate token", "error", err)
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	user, err := s.store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Account no longer exists.")
			return
		}
		s.logger.Error("get user", "error", err)
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}
