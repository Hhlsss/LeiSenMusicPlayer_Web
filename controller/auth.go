package controller

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"

	"MusicPlayerWeb/db"
)

type authReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
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
	// 使用Supabase Auth注册用户
	if err := db.RegisterUser(context.Background(), req.Account, req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// 生成UUID格式的用户ID
	userUUID := generateUUID(req.Account)

	// 创建用户个人资料（使用UUID格式）
	if err := db.CreateOrUpdateUserProfileByUUID(userUUID, req.Account, req.Account); err != nil {
		// 即使创建个人资料失败也继续，但记录错误
		fmt.Printf("创建用户个人资料失败: %v\n", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userUUID,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24小时
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "register ok",
		"nickname": req.Account,
		"user_id":  userUUID,
	})
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
	// 使用Supabase Auth登录用户
	if err := db.LoginUser(context.Background(), req.Account, req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	// 生成UUID格式的用户ID
	userUUID := generateUUID(req.Account)
	userProfile, err := db.GetUserProfileByEmail(req.Account)
	var nickname string

	if err != nil {
		// 如果用户个人资料不存在，创建新的个人资料
		nickname = req.Account
		if err := db.CreateOrUpdateUserProfileByUUID(userUUID, req.Account, nickname); err != nil {
			fmt.Printf("创建用户个人资料失败: %v", err)
		}
	} else {
		// 从数据库获取真实昵称
		if dbNickname, ok := userProfile["nickname"].(string); ok && dbNickname != "" {
			nickname = dbNickname
		} else {
			nickname = req.Account
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userUUID,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24小时
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "login ok",
		"nickname": nickname,
		"user_id":  userUUID,
	})
}

// HandleLogout 处理退出登录请求：POST /api/logout
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// 清除用户会话cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // 立即过期
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logout ok",
	})
}

// hashString 生成字符串的哈希值
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// generateUUID 生成UUID格式的用户ID
func generateUUID(account string) string {
	// 使用MD5哈希生成UUID格式的字符串
	hash := md5.Sum([]byte(account))
	uuid := hex.EncodeToString(hash[:])
	// 格式化为UUID格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:32])
}
