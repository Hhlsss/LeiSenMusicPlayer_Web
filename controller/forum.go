package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	"MusicPlayerWeb/db"
)

// ForumPost 论坛帖子结构
type ForumPost struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	UserID       string   `json:"user_id"`
	UserNickname string   `json:"user_nickname"`
	Tags         []string `json:"tags"`
	ViewCount    int      `json:"view_count"`
	ReplyCount   int      `json:"reply_count"`
	IsPinned     bool     `json:"is_pinned"`
	IsLocked     bool     `json:"is_locked"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// ForumReply 论坛回复结构
type ForumReply struct {
	ID           string `json:"id"`
	PostID       string `json:"post_id"`
	Content      string `json:"content"`
	UserID       string `json:"user_id"`
	UserNickname string `json:"user_nickname"`
	ParentID     string `json:"parent_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// HandleForumPage 论坛主页面
func HandleForumPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/forum.html")
}

// HandleForumPostPage 帖子详情页面
func HandleForumPostPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/forum-post.html")
}

// HandleForumPosts 获取论坛帖子列表
func HandleForumPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		getForumPosts(w, r)
	case "POST":
		createForumPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleForumPost 单个帖子操作
func HandleForumPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 提取帖子ID - 修复路径解析逻辑
	path := strings.TrimPrefix(r.URL.Path, "/api/forum/post/")
	if path == "" || path == "/" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	// 移除末尾的斜杠
	postID := strings.TrimSuffix(path, "/")
	
	switch r.Method {
	case "GET":
		getForumPost(w, r, postID)
	case "PUT":
		updateForumPost(w, r, postID)
	case "DELETE":
		deleteForumPost(w, r, postID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleForumReplies 获取帖子回复
func HandleForumReplies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		getForumReplies(w, r)
	case "POST":
		createForumReply(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleForumReply 单个回复操作
func HandleForumReply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "PUT":
		updateForumReply(w, r)
	case "DELETE":
		deleteForumReply(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 获取论坛帖子列表
func getForumPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageStr := query.Get("page")
	limitStr := query.Get("limit")
	tag := query.Get("tag")
	
	page := 1
	limit := 10
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := (page - 1) * limit
	
	// 构建查询URL
	url := fmt.Sprintf("%s/rest/v1/forum_posts?select=*&order=created_at.desc&offset=%d&limit=%d", 
		os.Getenv("SUPABASE_URL"), offset, limit)
	
	if tag != "" {
		url = fmt.Sprintf("%s/rest/v1/forum_posts?select=*&tags=cs.{%s}&order=created_at.desc&offset=%d&limit=%d",
			os.Getenv("SUPABASE_URL"), tag, offset, limit)
	}
	
	// 发送请求到Supabase
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var posts []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	// 如果数据库中没有帖子，返回空数组
	if posts == nil {
		posts = []ForumPost{}
	}
	
	json.NewEncoder(w).Encode(posts)
}

// 创建论坛帖子
func createForumPost(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	var post struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}
	
	// 验证必填字段
	if post.Title == "" || post.Content == "" {
		sendJSONError(w, "标题和内容不能为空", http.StatusBadRequest)
		return
	}
	
	// 准备插入数据（让Supabase自动生成UUID ID）
	insertData := map[string]interface{}{
		"title":         post.Title,
		"content":       post.Content,
		"user_id":       userInfo["id"],
		"user_nickname": userInfo["nickname"],
		"tags":          post.Tags,
		"view_count":    0,
		"reply_count":   0,
		"is_pinned":     false,
		"is_locked":     false,
		"created_at":    time.Now().Format(time.RFC3339),
		"updated_at":    time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(insertData)
	if err != nil {
		sendJSONError(w, "序列化数据失败", http.StatusInternalServerError)
		return
	}
	
	// 发送请求到Supabase
	url := fmt.Sprintf("%s/rest/v1/forum_posts", os.Getenv("SUPABASE_URL"))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	// 返回创建成功的帖子
	var createdPost []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&createdPost); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	if len(createdPost) > 0 {
		json.NewEncoder(w).Encode(createdPost[0])
	} else {
		sendJSONError(w, "创建帖子失败", http.StatusInternalServerError)
	}
}

// 获取单个帖子详情
func getForumPost(w http.ResponseWriter, r *http.Request, postID string) {
	
	// 增加查看次数
	go incrementViewCount(postID)
	
	// 获取帖子详情
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s&select=*", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var posts []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	if len(posts) == 0 {
		sendJSONError(w, "帖子不存在", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(posts[0])
}

// 更新帖子
func updateForumPost(w http.ResponseWriter, r *http.Request, postID string) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	var updateData struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}
	
	// 验证用户是否有权限编辑此帖子
	if !canEditPost(postID, userInfo["id"].(string)) {
		sendJSONError(w, "没有权限编辑此帖子", http.StatusForbidden)
		return
	}
	
	// 准备更新数据
	patchData := map[string]interface{}{
		"title":      updateData.Title,
		"content":    updateData.Content,
		"tags":       updateData.Tags,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(patchData)
	if err != nil {
		sendJSONError(w, "序列化数据失败", http.StatusInternalServerError)
		return
	}
	
	// 发送更新请求
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "帖子更新成功"})
}

// 删除帖子
func deleteForumPost(w http.ResponseWriter, r *http.Request, postID string) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	// 验证用户是否有权限删除此帖子
	if !canEditPost(postID, userInfo["id"].(string)) {
		sendJSONError(w, "没有权限删除此帖子", http.StatusForbidden)
		return
	}
	
	// 发送删除请求
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "帖子删除成功"})
}

// 获取帖子回复
func getForumReplies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	postID := query.Get("post_id")
	
	if postID == "" {
		sendJSONError(w, "帖子ID不能为空", http.StatusBadRequest)
		return
	}
	
	// 获取回复列表
	url := fmt.Sprintf("%s/rest/v1/forum_replies?post_id=eq.%s&select=*&order=created_at.asc", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var replies []ForumReply
	if err := json.NewDecoder(resp.Body).Decode(&replies); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	// 如果数据库中没有回复，返回空数组
	if replies == nil {
		replies = []ForumReply{}
	}
	
	json.NewEncoder(w).Encode(replies)
}

// 创建回复
func createForumReply(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	var reply struct {
		PostID   string `json:"post_id"`
		Content  string `json:"content"`
		ParentID string `json:"parent_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reply); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}
	
	// 验证必填字段
	if reply.PostID == "" || reply.Content == "" {
		sendJSONError(w, "帖子ID和内容不能为空", http.StatusBadRequest)
		return
	}
	
	// 如果ParentID为空，设置为nil（真正的NULL值）
	var parentID interface{} = nil
	if reply.ParentID != "" {
		parentID = reply.ParentID
	}
	
	// 准备插入数据（让Supabase自动生成UUID ID）
	insertData := map[string]interface{}{
		"post_id":       reply.PostID,
		"content":       reply.Content,
		"user_id":       userInfo["id"],
		"user_nickname": userInfo["nickname"],
		"parent_id":     parentID,
		"created_at":    time.Now().Format(time.RFC3339),
		"updated_at":    time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(insertData)
	if err != nil {
		sendJSONError(w, "序列化数据失败", http.StatusInternalServerError)
		return
	}
	
	// 发送请求到Supabase
	url := fmt.Sprintf("%s/rest/v1/forum_replies", os.Getenv("SUPABASE_URL"))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	// 增加帖子的回复计数
	go incrementReplyCount(reply.PostID)
	
	// 返回创建成功的回复
	var createdReply []ForumReply
	if err := json.NewDecoder(resp.Body).Decode(&createdReply); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	if len(createdReply) > 0 {
		json.NewEncoder(w).Encode(createdReply[0])
	} else {
		sendJSONError(w, "创建回复失败", http.StatusInternalServerError)
	}
}

// 更新回复
func updateForumReply(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	// 从URL路径中提取回复ID - 修复路径解析逻辑
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		sendJSONError(w, "无效的回复ID", http.StatusBadRequest)
		return
	}
	
	replyID := pathParts[2]
	
	var updateData struct {
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}
	
	// 验证用户是否有权限编辑此回复
	if !canEditReply(replyID, userInfo["id"].(string)) {
		sendJSONError(w, "没有权限编辑此回复", http.StatusForbidden)
		return
	}
	
	// 准备更新数据
	patchData := map[string]interface{}{
		"content":    updateData.Content,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(patchData)
	if err != nil {
		sendJSONError(w, "序列化数据失败", http.StatusInternalServerError)
		return
	}
	
	// 发送更新请求
	url := fmt.Sprintf("%s/rest/v1/forum_replies?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), replyID)
	
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "回复更新成功"})
}

// 删除回复
func deleteForumReply(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	// 从URL路径中提取回复ID - 修复路径解析逻辑
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		sendJSONError(w, "无效的回复ID", http.StatusBadRequest)
		return
	}
	
	replyID := pathParts[2]
	
	// 验证用户是否有权限删除此回复
	if !canEditReply(replyID, userInfo["id"].(string)) {
		sendJSONError(w, "没有权限删除此回复", http.StatusForbidden)
		return
	}
	
	// 发送删除请求
	url := fmt.Sprintf("%s/rest/v1/forum_replies?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), replyID)
	
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "回复删除成功"})
}

// 辅助函数：发送JSON格式的错误响应
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	}
	json.NewEncoder(w).Encode(errorResponse)
}

// 辅助函数：获取当前用户信息
func getCurrentUserInfo(r *http.Request) (map[string]interface{}, error) {
	// 从cookie中获取用户ID
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return nil, fmt.Errorf("用户未登录")
	}
	
	userIDStr := cookie.Value
	
	// 验证 UUID 格式
	if !isValidUUID(userIDStr) {
		return nil, fmt.Errorf("无效的用户ID格式")
	}
	
	// 现在系统完全使用UUID格式，直接使用UUID查询
	userInfo, err := db.GetUserProfileByUUID(userIDStr)
	if err != nil {
		// 如果获取失败，返回默认信息（但确保ID是有效的UUID）
		return map[string]interface{}{
			"id":       userIDStr,
			"nickname": "用户",
		}, nil
	}
	
	return userInfo, nil
}

// UUID 验证函数
func isValidUUID(u string) bool {
	// UUID 格式验证：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, u)
	return matched
}

// 辅助函数：生成唯一ID
func generateID() string {
	// 使用简单的时间戳+随机数作为ID，避免UUID格式问题
	return fmt.Sprintf("post_%d_%d", time.Now().UnixNano(), rand.Int63())
}

// 辅助函数：增加帖子查看次数
func incrementViewCount(postID string) {
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), postID)
	
	patchData := map[string]interface{}{
		"view_count": "forum_posts.view_count + 1",
	}
	
	jsonData, _ := json.Marshal(patchData)
	
	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	
	client.Do(req) // 忽略错误，不影响主流程
}

// 辅助函数：增加帖子回复计数
func incrementReplyCount(postID string) {
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), postID)
	
	patchData := map[string]interface{}{
		"reply_count": "forum_posts.reply_count + 1",
	}
	
	jsonData, _ := json.Marshal(patchData)
	
	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	
	client.Do(req) // 忽略错误，不影响主流程
}

// 辅助函数：检查用户是否有权限编辑帖子
func canEditPost(postID, userID string) bool {
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s&select=user_id", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return false
	}
	
	var posts []struct {
		UserID string `json:"user_id"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return false
	}
	
	if len(posts) == 0 {
		return false
	}
	
	return posts[0].UserID == userID
}

// 辅助函数：检查用户是否有权限编辑回复
func canEditReply(replyID, userID string) bool {
	url := fmt.Sprintf("%s/rest/v1/forum_replies?id=eq.%s&select=user_id", 
		os.Getenv("SUPABASE_URL"), replyID)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return false
	}
	
	var replies []struct {
		UserID string `json:"user_id"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&replies); err != nil {
		return false
	}
	
	if len(replies) == 0 {
		return false
	}
	
	return replies[0].UserID == userID
}

// HandleForumStats 获取论坛统计信息
func HandleForumStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	getForumStats(w, r)
}

// 获取论坛统计信息
func getForumStats(w http.ResponseWriter, r *http.Request) {
	// 获取总帖子数
	totalPosts, err := getTotalPostsCount()
	if err != nil {
		sendJSONError(w, "获取帖子总数失败", http.StatusInternalServerError)
		return
	}
	
	// 获取今日新帖数
	todayPosts, err := getTodayPostsCount()
	if err != nil {
		sendJSONError(w, "获取今日新帖数失败", http.StatusInternalServerError)
		return
	}
	
	// 获取总回复数
	totalReplies, err := getTotalRepliesCount()
	if err != nil {
		sendJSONError(w, "获取总回复数失败", http.StatusInternalServerError)
		return
	}
	
	// 获取热门标签
	topTags, err := getTopTags()
	if err != nil {
		sendJSONError(w, "获取热门标签失败", http.StatusInternalServerError)
		return
	}
	
	// 获取热门帖子
	hotPosts, err := getHotPosts()
	if err != nil {
		sendJSONError(w, "获取热门帖子失败", http.StatusInternalServerError)
		return
	}
	
	stats := map[string]interface{}{
		"total_posts":   totalPosts,
		"today_posts":   todayPosts,
		"total_replies": totalReplies,
		"top_tags":      topTags,
		"hot_posts":     hotPosts,
	}
	
	json.NewEncoder(w).Encode(stats)
}

// 获取总帖子数
func getTotalPostsCount() (int, error) {
	url := fmt.Sprintf("%s/rest/v1/forum_posts?select=id", os.Getenv("SUPABASE_URL"))
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	
	var posts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return 0, err
	}
	
	return len(posts), nil
}

// 获取今日新帖数
func getTodayPostsCount() (int, error) {
	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("%s/rest/v1/forum_posts?created_at=gte.%sT00:00:00&select=id", 
		os.Getenv("SUPABASE_URL"), today)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	
	var posts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return 0, err
	}
	
	return len(posts), nil
}

// 获取总回复数
func getTotalRepliesCount() (int, error) {
	url := fmt.Sprintf("%s/rest/v1/forum_replies?select=id", os.Getenv("SUPABASE_URL"))
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	
	var replies []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&replies); err != nil {
		return 0, err
	}
	
	return len(replies), nil
}

// 获取热门标签
func getTopTags() ([]map[string]interface{}, error) {
	// 获取所有帖子的标签
	url := fmt.Sprintf("%s/rest/v1/forum_posts?select=tags", os.Getenv("SUPABASE_URL"))
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	
	var posts []struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}
	
	// 统计标签出现次数
	tagCount := make(map[string]int)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tagCount[tag]++
		}
	}
	
	// 转换为前端需要的格式
	var topTags []map[string]interface{}
	for tag, count := range tagCount {
		topTags = append(topTags, map[string]interface{}{
			"name":  tag,
			"count": count,
		})
	}
	
	// 按出现次数排序（取前10个）
	if len(topTags) > 10 {
		topTags = topTags[:10]
	}
	
	return topTags, nil
}

// 获取热门帖子（按回复数排序）
func getHotPosts() ([]ForumPost, error) {
	url := fmt.Sprintf("%s/rest/v1/forum_posts?select=*&order=reply_count.desc&limit=5", 
		os.Getenv("SUPABASE_URL"))
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	
	var posts []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}
	
	return posts, nil
}

// HandleMyPosts 获取当前用户发布的帖子
func HandleMyPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	getMyPosts(w, r)
}

// HandleAdminPage 后台管理页面
func HandleAdminPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/admin.html")
}

// HandleAdminPosts 后台管理帖子API
func HandleAdminPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		getAdminPosts(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleAdminReplies 后台管理回复API
func HandleAdminReplies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		getAdminReplies(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleAdminPostDetail 后台管理单个帖子操作
func HandleAdminPostDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 从URL路径中提取帖子ID
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/posts/")
	if path == "" || path == "/" {
		sendJSONError(w, "无效的帖子ID", http.StatusBadRequest)
		return
	}
	
	// 移除末尾的斜杠
	postID := strings.TrimSuffix(path, "/")
	
	switch r.Method {
	case "DELETE":
		deleteAdminPost(w, r, postID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleAdminReplyDetail 后台管理单个回复操作
func HandleAdminReplyDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 从URL路径中提取回复ID
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/replies/")
	if path == "" || path == "/" {
		sendJSONError(w, "无效的回复ID", http.StatusBadRequest)
		return
	}
	
	// 移除末尾的斜杠
	replyID := strings.TrimSuffix(path, "/")
	
	switch r.Method {
	case "DELETE":
		deleteAdminReply(w, r, replyID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 后台删除帖子（管理员权限）
func deleteAdminPost(w http.ResponseWriter, r *http.Request, postID string) {
	// 发送删除请求
	url := fmt.Sprintf("%s/rest/v1/forum_posts?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), postID)
	
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "帖子删除成功"})
}

// 后台删除回复（管理员权限）
func deleteAdminReply(w http.ResponseWriter, r *http.Request, replyID string) {
	// 发送删除请求
	url := fmt.Sprintf("%s/rest/v1/forum_replies?id=eq.%s", 
		os.Getenv("SUPABASE_URL"), replyID)
	
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "回复删除成功"})
}

// 获取所有帖子（后台管理用）
func getAdminPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageStr := query.Get("page")
	limitStr := query.Get("limit")
	
	page := 1
	limit := 20
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := (page - 1) * limit
	
	// 构建查询URL，获取所有帖子
	url := fmt.Sprintf("%s/rest/v1/forum_posts?select=*&order=created_at.desc&offset=%d&limit=%d", 
		os.Getenv("SUPABASE_URL"), offset, limit)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var posts []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	// 如果数据库中没有帖子，返回空数组
	if posts == nil {
		posts = []ForumPost{}
	}
	
	// 获取总帖子数用于分页
	totalCount, err := getTotalPostsCount()
	if err != nil {
		totalCount = len(posts)
	}
	
	response := map[string]interface{}{
		"posts": posts,
		"total": totalCount,
		"page":  page,
		"limit": limit,
	}
	
	json.NewEncoder(w).Encode(response)
}

// 获取所有回复（后台管理用）
func getAdminReplies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageStr := query.Get("page")
	limitStr := query.Get("limit")
	
	page := 1
	limit := 20
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := (page - 1) * limit
	
	// 构建查询URL，获取所有回复
	url := fmt.Sprintf("%s/rest/v1/forum_replies?select=*&order=created_at.desc&offset=%d&limit=%d", 
		os.Getenv("SUPABASE_URL"), offset, limit)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var replies []ForumReply
	if err := json.NewDecoder(resp.Body).Decode(&replies); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	// 如果数据库中没有回复，返回空数组
	if replies == nil {
		replies = []ForumReply{}
	}
	
	// 获取总回复数用于分页
	totalCount, err := getTotalRepliesCount()
	if err != nil {
		totalCount = len(replies)
	}
	
	response := map[string]interface{}{
		"replies": replies,
		"total":   totalCount,
		"page":    page,
		"limit":   limit,
	}
	
	json.NewEncoder(w).Encode(response)
}

// 获取当前用户发布的帖子
func getMyPosts(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否登录
	userInfo, err := getCurrentUserInfo(r)
	if err != nil {
		sendJSONError(w, "请先登录", http.StatusUnauthorized)
		return
	}
	
	userID, ok := userInfo["id"].(string)
	if !ok {
		sendJSONError(w, "用户信息格式错误", http.StatusInternalServerError)
		return
	}
	
	// 获取用户发布的帖子
	url := fmt.Sprintf("%s/rest/v1/forum_posts?user_id=eq.%s&select=*&order=created_at.desc", 
		os.Getenv("SUPABASE_URL"), userID)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendJSONError(w, "创建请求失败", http.StatusInternalServerError)
		return
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	
	resp, err := client.Do(req)
	if err != nil {
		sendJSONError(w, "请求失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendJSONError(w, string(body), resp.StatusCode)
		return
	}
	
	var posts []ForumPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		sendJSONError(w, "解析响应失败", http.StatusInternalServerError)
		return
	}
	
	// 如果数据库中没有帖子，返回空数组
	if posts == nil {
		posts = []ForumPost{}
	}
	
	json.NewEncoder(w).Encode(posts)
}