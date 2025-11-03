package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"MusicPlayerWeb/service"
)

// HandleUploadMusic 处理音乐文件上传
func HandleUploadMusic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 解析multipart表单
	err = r.ParseMultipartForm(100 << 20) // 100MB限制
	if err != nil {
		writeErrUpload(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrUpload(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// 检查文件类型
	fileType := header.Header.Get("Content-Type")
	if !isValidMusicFileType(fileType, header.Filename) {
		writeErrUpload(w, http.StatusBadRequest, "invalid file type. only audio files are allowed")
		return
	}

	// 上传音乐文件
	musicFile, err := service.UploadMusicFile(header, userID)
	if err != nil {
		writeErrUpload(w, http.StatusInternalServerError, "upload failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "音乐文件上传成功",
		"music_file": musicFile,
	})
}

// HandleGetUserMusicFiles 获取用户上传的音乐文件列表
func HandleGetUserMusicFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 获取用户音乐文件列表
	musicFiles, err := service.GetUserMusicFiles(userID)
	if err != nil {
		writeErrUpload(w, http.StatusInternalServerError, "failed to get music files: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(musicFiles)
}

// HandleDeleteMusicFile 删除音乐文件
func HandleDeleteMusicFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 从URL路径中提取音乐文件ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeErrUpload(w, http.StatusBadRequest, "invalid music file id")
		return
	}
	musicFileID := pathParts[3]

	if musicFileID == "" {
		writeErrUpload(w, http.StatusBadRequest, "music file id required")
		return
	}

	// 删除音乐文件
	err = service.DeleteMusicFile(musicFileID, userID)
	if err != nil {
		writeErrUpload(w, http.StatusInternalServerError, "delete failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "音乐文件删除成功",
	})
}

// HandlePlayUploadedMusic 播放上传的音乐文件
func HandlePlayUploadedMusic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从URL参数中获取音乐文件ID
	musicFileID := r.URL.Query().Get("id")
	if musicFileID == "" {
		writeErrUpload(w, http.StatusBadRequest, "music file id required")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 获取用户音乐文件列表
	musicFiles, err := service.GetUserMusicFiles(userID)
	if err != nil {
		writeErrUpload(w, http.StatusInternalServerError, "failed to get music files: "+err.Error())
		return
	}

	// 查找指定的音乐文件
	var targetFile *service.MusicFile
	for _, file := range musicFiles {
		if file.ID == musicFileID {
			targetFile = &file
			break
		}
	}

	if targetFile == nil {
		writeErrUpload(w, http.StatusNotFound, "music file not found")
		return
	}

	// 获取音乐文件的URL
	musicURL := service.GetMusicFileURL(targetFile.StoragePath)

	// 重定向到音乐文件URL
	http.Redirect(w, r, musicURL, http.StatusFound)
}

// HandleCloudMusicList 获取云端音乐列表（包含本地和云端音乐）
func HandleCloudMusicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	var cloudMusic []service.MusicFile
	
	if err == nil {
		// 用户已登录，获取用户上传的云端音乐
		cloudMusic, err = service.GetUserMusicFiles(userID)
		if err != nil {
			// 获取云端音乐失败，但继续返回本地音乐
			cloudMusic = []service.MusicFile{}
		}
	} else {
		// 用户未登录，返回空云端音乐列表
		cloudMusic = []service.MusicFile{}
	}

	// 获取本地音乐（如果存在）
	localMusic, err := service.GetLocalMusicFiles()
	if err != nil {
		// 本地音乐获取失败不影响整体功能
		localMusic = []service.MusicFile{}
	}

	// 合并音乐列表
	allMusic := append(cloudMusic, localMusic...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cloud_music": cloudMusic,
		"local_music": localMusic,
		"total_count": len(allMusic),
	})
}

// HandleCloudMusicStream 流式播放云端音乐
func HandleCloudMusicStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从URL参数中获取音乐文件ID
	musicFileID := r.URL.Query().Get("id")
	if musicFileID == "" {
		writeErrUpload(w, http.StatusBadRequest, "music file id required")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 获取音乐文件信息
	musicFile, err := service.GetMusicFileByID(musicFileID, userID)
	if err != nil {
		writeErrUpload(w, http.StatusNotFound, "music file not found")
		return
	}

	// 设置流媒体响应头
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", musicFile.FileSize))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// 处理范围请求（支持断点续传）
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		service.HandleRangeRequest(w, r, musicFile)
		return
	}

	// 直接返回音乐文件URL
	http.Redirect(w, r, service.GetMusicFileURL(musicFile.StoragePath), http.StatusFound)
}

// isValidMusicFileType 检查文件类型是否为有效的音乐文件
func isValidMusicFileType(contentType string, filename string) bool {
	// 允许的音乐文件类型
	allowedTypes := map[string]bool{
		"audio/mpeg":       true, // MP3
		"audio/flac":       true, // FLAC
		"audio/wav":        true, // WAV
		"audio/ogg":        true, // OGG
		"audio/x-m4a":      true, // M4A
		"audio/aac":        true, // AAC
		"audio/x-ms-wma":   true, // WMA
		"audio/webm":       true, // WebM
		"application/octet-stream": true, // 二进制文件
	}

	// 检查Content-Type
	if allowedTypes[contentType] {
		return true
	}

	// 检查文件扩展名作为备用方案
	allowedExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".aac":  true,
		".wma":  true,
		".webm": true,
	}

	ext := strings.ToLower(filename[len(filename)-4:])
	if len(filename) > 4 && allowedExtensions[ext] {
		return true
	}

	// 检查3字符扩展名
	if len(filename) > 3 {
		ext3 := strings.ToLower(filename[len(filename)-3:])
		if allowedExtensions["."+ext3] {
			return true
		}
	}

	return false
}

// HandleUploadPage 处理上传页面
func HandleUploadPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取上传页面HTML文件
	content, err := os.ReadFile("web/upload.html")
	if err != nil {
		http.Error(w, "Upload page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

// HandleUploadStatus 处理上传状态查询
func HandleUploadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrUpload(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 检查用户是否已登录
	userID, err := service.GetCurrentUserID(r)
	if err != nil {
		writeErrUpload(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// 获取用户音乐文件统计
	musicFiles, err := service.GetUserMusicFiles(userID)
	if err != nil {
		writeErrUpload(w, http.StatusInternalServerError, "failed to get music files: "+err.Error())
		return
	}

	stats := map[string]interface{}{
		"total_files":    len(musicFiles),
		"total_size":     calculateTotalSize(musicFiles),
		"last_upload":    getLastUploadTime(musicFiles),
		"file_types":     getFileTypeStats(musicFiles),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 计算总文件大小
func calculateTotalSize(files []service.MusicFile) int64 {
	var total int64
	for _, file := range files {
		total += file.FileSize
	}
	return total
}

// 获取最后上传时间
func getLastUploadTime(files []service.MusicFile) string {
	if len(files) == 0 {
		return ""
	}
	return files[0].UploadedAt.Format("2006-01-02 15:04:05")
}

// 获取文件类型统计
func getFileTypeStats(files []service.MusicFile) map[string]int {
	stats := make(map[string]int)
	for _, file := range files {
		fileType := strings.ToLower(file.FileType)
		if fileType == "" {
			fileType = "unknown"
		}
		stats[fileType]++
	}
	return stats
}

// 辅助函数：写入错误响应
func writeErrUpload(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}