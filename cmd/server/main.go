package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"openai-server/internal/auth"
	"openai-server/internal/config"
	"openai-server/internal/proxy"
	"openai-server/internal/ratelimit"
)

type server struct {
	cfg         *config.Config
	users       map[string]struct{}
	blacklist   map[string]struct{}
	limiter     *ratelimit.Limiter
	proxyClient *proxy.Client
}

type accessKeyRequest struct {
	UserID string `json:"userId"`
}

type accessKeyResponse struct {
	AccessKey string `json:"access_key"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv := &server{
		cfg:         cfg,
		users:       toSet(cfg.Service.Users),
		blacklist:   toSet(cfg.Service.Blacklist),
		limiter:     ratelimit.New(cfg.RateLimit.Capacity, cfg.RateLimit.RefillIntervalSecond),
		proxyClient: proxy.New(cfg.OpenAI.TargetURL, cfg.OpenAI.APIKey),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/get_access_key", srv.handleGetAccessKey)
	mux.HandleFunc("/api/v1/chat/completions", srv.withAuth(srv.handleProxy))

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func (s *server) handleGetAccessKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var req accessKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if strings.TrimSpace(req.UserID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "userId is required"})
		return
	}

	if _, ok := s.users[req.UserID]; !ok {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "user not allowed"})
		return
	}

	if _, blocked := s.blacklist[req.UserID]; blocked {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "user in blacklist"})
		return
	}

	accessKey, err := auth.GenerateAccessKey(req.UserID, s.cfg.Service.SecretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "generate access key failed"})
		return
	}

	writeJSON(w, http.StatusOK, accessKeyResponse{AccessKey: accessKey})
}

func (s *server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessKey := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(accessKey), "bearer ") {
			accessKey = strings.TrimSpace(accessKey[7:])
		}
		if accessKey == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing access key"})
			return
		}

		userID, err := auth.ParseAccessKey(accessKey, s.cfg.Service.SecretKey)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid access key"})
			return
		}

		if _, ok := s.users[userID]; !ok {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "user not allowed"})
			return
		}
		if _, blocked := s.blacklist[userID]; blocked {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "user in blacklist"})
			return
		}

		if !s.limiter.Allow(userID) {
			writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "rate limit exceeded"})
			return
		}

		next(w, r)
	}
}

func (s *server) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	resp, err := s.proxyClient.Forward(r)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "proxy request failed"})
		return
	}

	proxy.CopyResponse(w, resp)
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		set[v] = struct{}{}
	}
	return set
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
