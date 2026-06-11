package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/store"
)

// Only images and common document formats are accepted.
var allowedContentTypes = []string{
	"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
	"application/pdf", "text/plain", "text/csv",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.ms-excel",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

func (s *Server) handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(s.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large",
			fmt.Sprintf("File exceeds the %d MB limit.", s.cfg.MaxUploadBytes>>20))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeValidationError(w, map[string]string{"file": "A file is required (multipart field \"file\")."})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !slices.Contains(allowedContentTypes, contentType) {
		writeValidationError(w, map[string]string{"file": "Unsupported file type. Allowed: images, PDF, text, Word, Excel."})
		return
	}

	// Store under a random name to avoid path traversal and collisions.
	storedName := uuid.NewString() + sanitizeExt(filepath.Ext(header.Filename))
	dst, err := os.Create(filepath.Join(s.cfg.UploadDir, storedName))
	if err != nil {
		s.logger.Error("create upload file", "error", err)
		writeInternalError(w)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		s.logger.Error("write upload file", "error", err)
		writeInternalError(w)
		return
	}

	attachment := &models.Attachment{
		TaskID:      task.ID,
		FileName:    filepath.Base(header.Filename),
		StoredName:  storedName,
		ContentType: contentType,
		SizeBytes:   size,
	}
	created, err := s.store.CreateAttachment(r.Context(), attachment)
	if err != nil {
		_ = os.Remove(filepath.Join(s.cfg.UploadDir, storedName))
		s.logger.Error("create attachment", "error", err)
		writeInternalError(w)
		return
	}

	s.recordActivity(r, task.ID, "attachment_added", fmt.Sprintf("Attached file %q", created.FileName))
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func sanitizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if len(ext) > 10 || strings.ContainsAny(ext, "/\\") {
		return ""
	}
	return ext
}

func (s *Server) handleListAttachments(w http.ResponseWriter, r *http.Request) {
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}

	attachments, err := s.store.ListAttachments(r.Context(), task.ID)
	if err != nil {
		s.logger.Error("list attachments", "error", err)
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": attachments})
}

// loadAttachmentAuthorized resolves an attachment and verifies the requester
// may access its parent task.
func (s *Server) loadAttachmentAuthorized(w http.ResponseWriter, r *http.Request) *models.Attachment {
	claims := claimsFrom(r.Context())
	id := chi.URLParam(r, "id")

	attachment, err := s.store.GetAttachment(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Attachment not found.")
			return nil
		}
		s.logger.Error("get attachment", "error", err)
		writeInternalError(w)
		return nil
	}

	task, err := s.store.GetTask(r.Context(), attachment.TaskID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Attachment not found.")
		return nil
	}
	if task.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		writeError(w, http.StatusNotFound, "not_found", "Attachment not found.")
		return nil
	}
	return attachment
}

func (s *Server) handleDownloadAttachment(w http.ResponseWriter, r *http.Request) {
	attachment := s.loadAttachmentAuthorized(w, r)
	if attachment == nil {
		return
	}

	path := filepath.Join(s.cfg.UploadDir, filepath.Base(attachment.StoredName))
	f, err := os.Open(path)
	if err != nil {
		s.logger.Error("open attachment file", "error", err)
		writeError(w, http.StatusNotFound, "not_found", "Attachment file is missing.")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.FileName))
	_, _ = io.Copy(w, f)
}

func (s *Server) handleDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	attachment := s.loadAttachmentAuthorized(w, r)
	if attachment == nil {
		return
	}

	if err := s.store.DeleteAttachment(r.Context(), attachment.ID); err != nil {
		s.logger.Error("delete attachment", "error", err)
		writeInternalError(w)
		return
	}
	_ = os.Remove(filepath.Join(s.cfg.UploadDir, filepath.Base(attachment.StoredName)))

	s.recordActivity(r, attachment.TaskID, "attachment_removed", fmt.Sprintf("Removed file %q", attachment.FileName))
	w.WriteHeader(http.StatusNoContent)
}
