package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/impossibleclone/fserver/internal/auth"
	"github.com/impossibleclone/fserver/internal/config"
	"github.com/impossibleclone/fserver/internal/handler"
	"github.com/impossibleclone/fserver/internal/security"
	"github.com/impossibleclone/fserver/internal/storage"
)

type FileServer struct {
	cfg           *config.Config
	httpServer    *http.Server
	authenticator auth.Authenticator
	logger        *log.Logger
}

func NewFileServer(cfg *config.Config, authenticator auth.Authenticator) *FileServer {
	logFile, _ := os.OpenFile("server_audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	logger := log.New(ioMultiWriter(os.Stdout, logFile), "[AUDIT] ", log.Ldate|log.Ltime)

	return &FileServer{
		cfg:           cfg,
		authenticator: authenticator,
		logger:        logger,
	}
}

func ioMultiWriter(writers ...interface{}) *multiWriter {
	return &multiWriter{writers: writers}
}
type multiWriter struct { writers []interface{} }
func (t *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		if writer, ok := w.(interface{ Write([]byte) (int, error) }); ok {
			writer.Write(p)
		}
	}
	return len(p), nil
}

type responseWriter struct {
	http.ResponseWriter
	status int
}
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *FileServer) Start() error {
	if err := security.EnsureCerts("cert.pem", "key.pem"); err != nil {
		s.logger.Printf("Failed to generate TLS certs: %v", err)
		return err
	}

	mux := http.NewServeMux()
	
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			u, p, ok := r.BasicAuth()
			
			if !ok || s.authenticator.Authenticate(u, p) != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				s.logger.Printf("UNAUTHORIZED ACCESS ATTEMPT from IP: %s to %s", r.RemoteAddr, r.URL.Path)
				return
			}
			
			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)
			
			s.logger.Printf("User: '%s' | Action: %s | Target: %s | Status: %d | Time: %v | IP: %s", 
				u, r.Method, r.URL.Path, rw.status, time.Since(start), r.RemoteAddr)
		}
	}

	// strictAdminMiddleware requires the user to literally be "admin" (or match cfg.Username)
	strictAdminMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			u, _, _ := r.BasicAuth()
			if u != s.cfg.Username {
				http.Error(w, "Forbidden: Admins Only", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	vfs := storage.NewLocalFS(s.cfg.StorageDir)
	h := handler.NewHandler(vfs, s.authenticator)
	
	// Public (Authenticated) Endpoints
	mux.HandleFunc("/", authMiddleware(h.HandleWebUI))
	mux.HandleFunc("/whoami", authMiddleware(h.HandleWhoAmI))
	mux.HandleFunc("/upload", authMiddleware(h.HandleUpload))
	mux.HandleFunc("/download/", authMiddleware(h.HandleDownload))
	mux.HandleFunc("/list", authMiddleware(h.HandleList))
	mux.HandleFunc("/delete/", authMiddleware(h.HandleDelete))

	// Admin Only Endpoints
	mux.HandleFunc("/admin", strictAdminMiddleware(h.HandleAdminUI))
	mux.HandleFunc("/admin/users", strictAdminMiddleware(h.HandleAdminUsers))

	addr := fmt.Sprintf(":%s", s.cfg.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.Printf("Server starting on HTTPS (TLS) port %s", s.cfg.Port)
	return s.httpServer.ListenAndServeTLS("cert.pem", "key.pem")
}

func (s *FileServer) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		s.logger.Printf("Server shutting down...")
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
