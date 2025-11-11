package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"strconv"
	"strings"
)

// AdminMiddleware 管理员权限验证中间件
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取用户ID
		userID, err := getCurrentUserID(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
			return
		}

		// 检查用户是否为管理员
		isAdmin, err := checkUserIsAdmin(userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
			return
		}

		if !isAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
			return
		}

		// 用户是管理员，继续处理请求
		next(w, r)
	}
}

// getCurrentUserID 从请求中获取当前用户ID
func getCurrentUserID(r *http.Request) (string, error) {
	// 首先尝试从cookie获取
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// 然后尝试从Authorization头获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			// 解析JWT sub为用户ID
			if sub := parseJWTSub(token); sub != "" {
				return sub, nil
			}
			return token, nil
		}
	}

	// 最后尝试从查询参数获取
	userID := r.URL.Query().Get("user_id")
	if userID != "" {
		return userID, nil
	}

	return "", fmt.Errorf("未找到用户ID")
}

// getUserToken 从请求中获取用户认证token
func getUserToken(r *http.Request) string {
	// 首先尝试从Authorization头获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 然后尝试从cookie获取
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 最后尝试从查询参数获取
	token := r.URL.Query().Get("access_token")
	if token != "" {
		return token
	}

	return ""
}

// checkUserIsAdmin 检查用户是否为管理员
func checkUserIsAdmin(userID string) (bool, error) {
	// 使用HTTP客户端直接调用Supabase REST API检查管理员权限
	httpClient := &http.Client{}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%s&select=is_admin", os.Getenv("SUPABASE_URL"), userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result) > 0 {
			if isAdmin, ok := result[0]["is_admin"].(bool); ok {
				return isAdmin, nil
			}
		}
	}

	return false, nil
}

// HandleAdminDashboard 管理员仪表板
func HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// 获取统计信息
	stats, err := getAdminStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取统计信息失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "管理员仪表板",
		"stats":   stats,
	})
}

// getAdminStats 获取管理员统计信息
func getAdminStats() (map[string]interface{}, error) {
	httpClient := &http.Client{}

	// 获取用户总数
	usersUrl := fmt.Sprintf("%s/rest/v1/user_profiles?select=count", os.Getenv("SUPABASE_URL"))
	usersReq, err := http.NewRequest("GET", usersUrl, nil)
	if err != nil {
		return nil, err
	}
	usersReq.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	usersReq.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	usersReq.Header.Set("Content-Type", "application/json")
	usersReq.Header.Set("Prefer", "count=exact")

	usersResp, err := httpClient.Do(usersReq)
	if err != nil {
		return nil, err
	}
	defer usersResp.Body.Close()

	userCount := 0
	if usersResp.StatusCode == http.StatusOK {
		// 从Content-Range头获取总数
		contentRange := usersResp.Header.Get("Content-Range")
		if contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				fmt.Sscanf(parts[1], "%d", &userCount)
			}
		}
	}

	// 获取帖子总数
	postsUrl := fmt.Sprintf("%s/rest/v1/forum_posts?select=count", os.Getenv("SUPABASE_URL"))
	postsReq, err := http.NewRequest("GET", postsUrl, nil)
	if err != nil {
		return nil, err
	}
	postsReq.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	postsReq.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	postsReq.Header.Set("Content-Type", "application/json")
	postsReq.Header.Set("Prefer", "count=exact")

	postsResp, err := httpClient.Do(postsReq)
	if err != nil {
		return nil, err
	}
	defer postsResp.Body.Close()

	postCount := 0
	if postsResp.StatusCode == http.StatusOK {
		contentRange := postsResp.Header.Get("Content-Range")
		if contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				fmt.Sscanf(parts[1], "%d", &postCount)
			}
		}
	}

	return map[string]interface{}{
		"total_users": userCount,
		"total_posts": postCount,
		"admin_users": 1, // 暂时硬编码，后续可以从数据库获取
	}, nil
}

// HandleAdminUsers 管理员用户管理
func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getAdminUsers(w, r)
	case "PUT":
		updateAdminUser(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// getAdminUsers 获取所有用户列表（管理员权限）
func getAdminUsers(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否为管理员
	userID, err := getCurrentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	// 验证用户是否为管理员
	isAdmin, err := checkUserIsAdmin(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
		return
	}

	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
		return
	}

	httpClient := &http.Client{}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?select=*", os.Getenv("SUPABASE_URL"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建请求失败"})
		return
	}

	// 管理员已校验，使用 anon key 服务端直连，避免前端令牌缺失导致 401
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var users []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&users); err == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"users": users,
			})
			return
		}
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取用户列表失败"})
}

// updateAdminUser 更新用户信息（管理员权限）
func updateAdminUser(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否为管理员
	userID, err := getCurrentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	// 验证用户是否为管理员
	isAdmin, err := checkUserIsAdmin(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
		return
	}

	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
		return
	}

	var updateData struct {
		UserID   string `json:"user_id"`
		IsAdmin  *bool  `json:"is_admin"`
		Nickname string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的JSON数据"})
		return
	}

	if updateData.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户ID不能为空"})
		return
	}

	httpClient := &http.Client{}

	patchData := map[string]interface{}{
		"updated_at": "now()",
	}

	if updateData.IsAdmin != nil {
		patchData["is_admin"] = *updateData.IsAdmin
	}

	if updateData.Nickname != "" {
		patchData["nickname"] = updateData.Nickname
	}

	jsonData, err := json.Marshal(patchData)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "序列化数据失败"})
		return
	}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%s", os.Getenv("SUPABASE_URL"), updateData.UserID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建请求失败"})
		return
	}

	// 管理员已校验，使用 anon key 服务端直连
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		writeJSON(w, http.StatusOK, map[string]string{"message": "用户信息更新成功"})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新用户信息失败"})
}

// HandleAdminPosts 管理员帖子管理
func HandleAdminPosts(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/admin/posts/") {
		// 处理带ID的路径，如 /api/admin/posts/123
		postID := strings.TrimPrefix(path, "/api/admin/posts/")
		if postID != "" {
			switch r.Method {
			case "DELETE":
				deleteAdminPostByPath(w, r, postID)
			default:
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			}
			return
		}
	}

	// 处理不带ID的路径
	switch r.Method {
	case "GET":
		getAdminPosts(w, r)
	case "DELETE":
		deleteAdminPost(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// getAdminPosts 获取所有帖子列表（管理员权限）
func getAdminPosts(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否为管理员
	userID, err := getCurrentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	// 验证用户是否为管理员
	isAdmin, err := checkUserIsAdmin(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
		return
	}

	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
		return
	}

	httpClient := &http.Client{}

	// 获取查询参数
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	// 构建查询URL
	baseURL := fmt.Sprintf("%s/rest/v1/forum_posts?select=*", os.Getenv("SUPABASE_URL"))

	// 添加排序
	baseURL += "&order=created_at.desc"

	// 添加分页
	if limit != "" {
		baseURL += fmt.Sprintf("&limit=%s", limit)
	}
	if page != "" && limit != "" {
		pageNum := 1
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
		limitNum := 10
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitNum = l
		}
		offset := (pageNum - 1) * limitNum
		baseURL += fmt.Sprintf("&offset=%d", offset)
	}

	// 添加搜索条件
	if search != "" {
		esc := urlpkg.QueryEscape("*" + search + "*")
		baseURL += fmt.Sprintf("&or=(title.ilike.%s,content.ilike.%s,user_nickname.ilike.%s)", esc, esc, esc)
	}

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建请求失败"})
		return
	}

	// 优先转发用户JWT；若无JWT，降级为 anon key（需RLS允许只读）
	token := getUserToken(r)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "count=exact")

	resp, err := httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var posts []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&posts); err == nil {
			// 获取总数
			total := 0
			contentRange := resp.Header.Get("Content-Range")
			if contentRange != "" {
				parts := strings.Split(contentRange, "/")
				if len(parts) == 2 {
					fmt.Sscanf(parts[1], "%d", &total)
				}
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"posts": posts,
				"total": total,
			})
			return
		}
	}

	// 读取错误响应
	body, _ := io.ReadAll(resp.Body)
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": fmt.Sprintf("获取帖子列表失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(body)),
	})
}

// deleteAdminPost 删除帖子（管理员权限）
func deleteAdminPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "帖子ID不能为空"})
		return
	}

	// 检查用户是否为管理员
	userID, err := getCurrentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	// 验证用户是否为管理员
	isAdmin, err := checkUserIsAdmin(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
		return
	}

	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
		return
	}

	httpClient := &http.Client{}

	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", os.Getenv("SUPABASE_URL"), postID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建请求失败"})
		return
	}

	// 管理员已校验，使用 service role key 以绕过 RLS（从 env 或内置后备获取）
	serviceKey := getSupabaseServiceKey()
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应体并校验是否真的删除
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		var deleted []map[string]interface{}
		_ = json.Unmarshal(body, &deleted)
		if len(deleted) == 0 {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "未删除任何记录，可能是RLS或ID不存在"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "帖子删除成功"})
		return
	}

	// 输出详细的错误信息
	fmt.Printf("删除帖子失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(body))

	// 如果认证失败，提示用户重新登录
	if resp.StatusCode == http.StatusForbidden {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "权限不足，请确保您已登录且具有管理员权限"})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("删除帖子失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(body))})
}

// deleteAdminPostByPath 通过路径参数删除帖子（管理员权限）
func deleteAdminPostByPath(w http.ResponseWriter, r *http.Request, postID string) {
	if postID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "帖子ID不能为空"})
		return
	}

	// 检查用户是否为管理员
	userID, err := getCurrentUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	// 验证用户是否为管理员
	isAdmin, err := checkUserIsAdmin(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "权限验证失败"})
		return
	}

	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "需要管理员权限"})
		return
	}

	httpClient := &http.Client{}

	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", os.Getenv("SUPABASE_URL"), postID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建请求失败"})
		return
	}

	// 管理员已校验，使用 service role key 以绕过 RLS（从 env 或内置后备获取）
	serviceKey := getSupabaseServiceKey()
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应体并校验是否真的删除
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		var deleted []map[string]interface{}
		_ = json.Unmarshal(body, &deleted)
		if len(deleted) == 0 {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "未删除任何记录，可能是RLS或ID不存在"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "帖子删除成功"})
		return
	}

	// 输出详细的错误信息
	fmt.Printf("删除帖子失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(body))

	// 如果认证失败，提示用户重新登录
	if resp.StatusCode == http.StatusForbidden {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "权限不足，请确保您已登录且具有管理员权限"})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("删除帖子失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(body))})
}

// writeJSON 辅助函数：写入JSON响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// 解析JWT的sub（仅为获取用户ID用途，不在此校验签名）
func parseJWTSub(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}
	payloadSeg := parts[1]
	b, err := base64.RawURLEncoding.DecodeString(payloadSeg)
	if err != nil {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(b, &payload); err != nil {
		return ""
	}
	if sub, ok := payload["sub"].(string); ok {
		return sub
	}
	return ""
}

// getSupabaseServiceKey 返回服务端使用的 service role key（优先环境变量，其次内置常量）
func getSupabaseServiceKey() string {
	if v := os.Getenv("SUPABASE_SERVICE_ROLE_KEY"); v != "" {
		return v
	}
	// 用户提供的 service role key 作为后备值
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdibG5wenN0ZGpudmNsaWpqcGJrIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2MDQ2ODM2MSwiZXhwIjoyMDc2MDQ0MzYxfQ.eB9ZA8Ymqa4nxzorUskMLT1V4jP8AlGX4Soic2g8DYQ"
}
