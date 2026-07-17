package handler

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/impossibleclone/fserver/internal/auth"
	"github.com/impossibleclone/fserver/internal/storage"
)

//go:embed web/index.html
var webUITemplate []byte

//go:embed web/admin.html
var adminUITemplate []byte

type Handler struct {
	vfs  storage.VFS
	auth auth.Authenticator
}

func NewHandler(vfs storage.VFS, authenticator auth.Authenticator) *Handler {
	return &Handler{
		vfs:  vfs,
		auth: authenticator,
	}
}

func (h *Handler) HandleWhoAmI(w http.ResponseWriter, r *http.Request) {
	u, _, _ := r.BasicAuth()
	w.Write([]byte(u))
}

func (h *Handler) HandleWebUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(webUITemplate)
}

func (h *Handler) HandleAdminUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(adminUITemplate)
}

func (h *Handler) HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		users := h.auth.GetUsers()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		return
	}
	
	if r.Method == http.MethodPost {
		u := r.URL.Query().Get("username")
		p := r.URL.Query().Get("password")
		if err := h.auth.AddUser(u, p); err != nil {
			http.Error(w, "Failed to add user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method == http.MethodDelete {
		u := r.URL.Query().Get("username")
		if u == "admin" {
			http.Error(w, "Cannot delete admin", http.StatusBadRequest)
			return
		}
		h.auth.RemoveUser(u)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	overwrite := r.URL.Query().Get("overwrite") == "true"

	if err := h.vfs.SaveFile(filename, file, overwrite); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := filepath.Base(r.URL.Path)
	file, err := h.vfs.GetFile(filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	http.ServeContent(w, r, filename, time.Now(), file)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := filepath.Base(r.URL.Path)
	if err := h.vfs.DeleteFile(filename); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	files, err := h.vfs.ListFiles()
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
