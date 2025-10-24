package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"MusicPlayerWeb/service"
)

type authReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// HandleRegister 处理注册请求：POST /api/register
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Account == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account and password required"})
		return
	}
	if err := service.RegisterUser(context.Background(), req.Account, req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "register ok"})
}

// HandleLogin 处理登录请求：POST /api/login
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Account == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account and password required"})
		return
	}
	if err := service.LoginUser(context.Background(), req.Account, req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "login ok",
		"nickname": req.Account,
	})
}
