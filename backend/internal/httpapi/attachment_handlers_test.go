package httpapi

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/milann/taskflow/internal/models"
)

func uploadFile(t *testing.T, h http.Handler, taskID, token, fileName, contentType string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+fileName+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/"+taskID+"/attachments", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAttachments(t *testing.T) {
	t.Run("uploads, lists, downloads and deletes a file", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := uploadFile(t, h, task.ID, token, "notes.txt", "text/plain", []byte("hello world"))
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		created := decodeBody[map[string]models.Attachment](t, rec)["data"]
		if created.FileName != "notes.txt" || created.SizeBytes != 11 {
			t.Errorf("unexpected attachment %+v", created)
		}

		listRec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID+"/attachments", token, nil)
		list := decodeBody[map[string][]models.Attachment](t, listRec)["data"]
		if len(list) != 1 {
			t.Fatalf("expected 1 attachment, got %d", len(list))
		}

		dlRec := doRequest(t, h, http.MethodGet, "/api/attachments/"+created.ID+"/download", token, nil)
		if dlRec.Code != http.StatusOK || dlRec.Body.String() != "hello world" {
			t.Errorf("download failed: %d %q", dlRec.Code, dlRec.Body.String())
		}

		delRec := doRequest(t, h, http.MethodDelete, "/api/attachments/"+created.ID, token, nil)
		if delRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", delRec.Code)
		}
	})

	t.Run("rejects disallowed content type", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := uploadFile(t, h, task.ID, token, "app.exe", "application/x-msdownload", []byte{0x4D, 0x5A})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("rejects missing file field", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		_ = writer.WriteField("other", "value")
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/tasks/"+task.ID+"/attachments", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("cannot access attachments of other users' tasks", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, aliceToken := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, bobToken := createTestUser(t, st, "bob@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := uploadFile(t, h, task.ID, aliceToken, "notes.txt", "text/plain", []byte("secret"))
		created := decodeBody[map[string]models.Attachment](t, rec)["data"]

		if rec := uploadFile(t, h, task.ID, bobToken, "x.txt", "text/plain", []byte("x")); rec.Code != http.StatusNotFound {
			t.Errorf("expected upload to other's task to 404, got %d", rec.Code)
		}
		if rec := doRequest(t, h, http.MethodGet, "/api/attachments/"+created.ID+"/download", bobToken, nil); rec.Code != http.StatusNotFound {
			t.Errorf("expected download of other's attachment to 404, got %d", rec.Code)
		}
		if rec := doRequest(t, h, http.MethodDelete, "/api/attachments/"+created.ID, bobToken, nil); rec.Code != http.StatusNotFound {
			t.Errorf("expected delete of other's attachment to 404, got %d", rec.Code)
		}
	})
}
