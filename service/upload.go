package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

// MusicFile 表示上传到数据库的音乐文件
type MusicFile struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Album       string    `json:"album"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	FileType    string    `json:"file_type"`
	StoragePath string    `json:"storage_path"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UserID      string    `json:"user_id"`
}

// UploadMusicFile 上传音乐文件到Supabase存储并保存元数据到数据库
func UploadMusicFile(fileHeader *multipart.FileHeader, userUUID string) (*MusicFile, error) {
	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 提取音乐元数据
	metadata, err := extractMusicMetadata(fileBytes, fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("提取元数据失败: %v", err)
	}

	// 上传文件到Supabase存储
	storagePath, err := uploadToSupabaseStorage(fileBytes, fileHeader.Filename, userUUID)
	if err != nil {
		return nil, fmt.Errorf("上传到存储失败: %v", err)
	}

	// 保存音乐元数据到数据库
	musicFile, err := saveMusicMetadata(metadata, fileHeader, storagePath, userUUID)
	if err != nil {
		// 如果保存元数据失败，尝试删除已上传的文件
		deleteFromSupabaseStorage(storagePath)
		return nil, fmt.Errorf("保存元数据失败: %v", err)
	}

	return musicFile, nil
}

// extractMusicMetadata 从文件内容中提取音乐元数据
func extractMusicMetadata(fileBytes []byte, filename string) (map[string]string, error) {
	metadata := map[string]string{
		"title":  strings.TrimSuffix(filename, filepath.Ext(filename)),
		"artist": "未知艺术家",
		"album":  "未知专辑",
	}

	// 使用tag库读取音乐文件元数据
	reader := bytes.NewReader(fileBytes)
	m, err := tag.ReadFrom(reader)
	if err != nil {
		// 如果无法读取元数据，使用文件名作为标题
		return metadata, nil
	}

	if m != nil {
		if title := m.Title(); title != "" {
			metadata["title"] = title
		}
		if artist := m.Artist(); artist != "" {
			metadata["artist"] = artist
		}
		if album := m.Album(); album != "" {
			metadata["album"] = album
		}
	}

	return metadata, nil
}

// uploadToSupabaseStorage 上传文件到Supabase存储桶
func uploadToSupabaseStorage(fileBytes []byte, filename string, userUUID string) (string, error) {
	// 生成唯一的存储路径
	timestamp := time.Now().Unix()
	storagePath := fmt.Sprintf("music/%s/%d_%s", userUUID, timestamp, filename)

	// 使用Supabase Storage API上传文件
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/storage/v1/object/music/%s", os.Getenv("SUPABASE_URL"), storagePath)

	req, err := http.NewRequest("POST", url, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("上传失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return storagePath, nil
}

// deleteFromSupabaseStorage 从Supabase存储桶删除文件
func deleteFromSupabaseStorage(storagePath string) error {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/storage/v1/object/music/%s", os.Getenv("SUPABASE_URL"), storagePath)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建删除请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("删除请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("删除失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// saveMusicMetadata 保存音乐元数据到数据库
func saveMusicMetadata(metadata map[string]string, fileHeader *multipart.FileHeader, storagePath string, userUUID string) (*MusicFile, error) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/music_files", os.Getenv("SUPABASE_URL"))

	musicFile := &MusicFile{
		ID:          fmt.Sprintf("%s_%d", userUUID, time.Now().Unix()),
		Title:       metadata["title"],
		Artist:      metadata["artist"],
		Album:       metadata["album"],
		FileName:    fileHeader.Filename,
		FileSize:    fileHeader.Size,
		FileType:    strings.ToLower(filepath.Ext(fileHeader.Filename)),
		StoragePath: storagePath,
		UploadedAt:  time.Now(),
		UserID:      userUUID,
	}

	insertData := map[string]interface{}{
		"id":           musicFile.ID,
		"title":        musicFile.Title,
		"artist":       musicFile.Artist,
		"album":        musicFile.Album,
		"file_name":    musicFile.FileName,
		"file_size":    musicFile.FileSize,
		"file_type":    musicFile.FileType,
		"storage_path": musicFile.StoragePath,
		"uploaded_at":  musicFile.UploadedAt.Format(time.RFC3339),
		"user_id":      musicFile.UserID,
	}

	jsonData, err := json.Marshal(insertData)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// 如果表不存在，尝试创建表
		if resp.StatusCode == http.StatusNotFound {
			if err := createMusicFilesTable(); err != nil {
				return nil, fmt.Errorf("创建音乐文件表失败: %v", err)
			}
			// 重新尝试插入数据
			return saveMusicMetadata(metadata, fileHeader, storagePath, userUUID)
		}

		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("保存元数据失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return musicFile, nil
}

// createMusicFilesTable 创建音乐文件表
func createMusicFilesTable() error {
	httpClient := &http.Client{}

	// 创建表的SQL语句
	sql := `CREATE TABLE IF NOT EXISTS music_files (
		id VARCHAR PRIMARY KEY,
		title VARCHAR NOT NULL,
		artist VARCHAR NOT NULL,
		album VARCHAR,
		file_name VARCHAR NOT NULL,
		file_size BIGINT NOT NULL,
		file_type VARCHAR NOT NULL,
		storage_path VARCHAR NOT NULL,
		uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		user_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES auth.users(id)
	)`

	requestData := map[string]interface{}{
		"query": sql,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化SQL语句失败: %v", err)
	}

	url := fmt.Sprintf("%s/rest/v1/", os.Getenv("SUPABASE_URL"))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建表失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetUserMusicFiles 获取用户上传的音乐文件列表
func GetUserMusicFiles(userUUID string) ([]MusicFile, error) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/music_files?user_id=eq.%s&order=uploaded_at.desc", os.Getenv("SUPABASE_URL"), userUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果表不存在，返回空列表
		if resp.StatusCode == http.StatusNotFound {
			return []MusicFile{}, nil
		}
		return nil, fmt.Errorf("获取音乐文件失败，状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	var musicFiles []MusicFile
	for _, item := range result {
		musicFile := MusicFile{
			ID:          getStringFromMapUpload(item, "id", ""),
			Title:       getStringFromMapUpload(item, "title", ""),
			Artist:      getStringFromMapUpload(item, "artist", ""),
			Album:       getStringFromMapUpload(item, "album", ""),
			FileName:    getStringFromMapUpload(item, "file_name", ""),
			FileSize:    getInt64FromMapUpload(item, "file_size", 0),
			FileType:    getStringFromMapUpload(item, "file_type", ""),
			StoragePath: getStringFromMapUpload(item, "storage_path", ""),
			UserID:      getStringFromMapUpload(item, "user_id", ""),
		}

		// 解析上传时间
		if uploadedAtStr := getStringFromMap(item, "uploaded_at", ""); uploadedAtStr != "" {
			if uploadedAt, err := time.Parse(time.RFC3339, uploadedAtStr); err == nil {
				musicFile.UploadedAt = uploadedAt
			}
		}

		musicFiles = append(musicFiles, musicFile)
	}

	return musicFiles, nil
}

// GetMusicFileURL 获取音乐文件的访问URL
func GetMusicFileURL(storagePath string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/music/%s", os.Getenv("SUPABASE_URL"), storagePath)
}

// GetMusicFileByID 根据ID获取音乐文件信息
func GetMusicFileByID(musicFileID string, userUUID string) (*MusicFile, error) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/music_files?id=eq.%s&user_id=eq.%s", os.Getenv("SUPABASE_URL"), musicFileID, userUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("音乐文件不存在，状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result) == 0 {
		return nil, fmt.Errorf("音乐文件不存在")
	}

	item := result[0]
	musicFile := &MusicFile{
		ID:          getStringFromMapUpload(item, "id", ""),
		Title:       getStringFromMapUpload(item, "title", ""),
		Artist:      getStringFromMapUpload(item, "artist", ""),
		Album:       getStringFromMapUpload(item, "album", ""),
		FileName:    getStringFromMapUpload(item, "file_name", ""),
		FileSize:    getInt64FromMapUpload(item, "file_size", 0),
		FileType:    getStringFromMapUpload(item, "file_type", ""),
		StoragePath: getStringFromMapUpload(item, "storage_path", ""),
		UserID:      getStringFromMapUpload(item, "user_id", ""),
	}

	// 解析上传时间
	if uploadedAtStr := getStringFromMapUpload(item, "uploaded_at", ""); uploadedAtStr != "" {
		if uploadedAt, err := time.Parse(time.RFC3339, uploadedAtStr); err == nil {
			musicFile.UploadedAt = uploadedAt
		}
	}

	return musicFile, nil
}

// GetLocalMusicFiles 获取本地音乐文件列表
func GetLocalMusicFiles() ([]MusicFile, error) {
	// 这里实现获取本地音乐文件的逻辑
	// 暂时返回空列表，后续可以根据实际需求实现
	return []MusicFile{}, nil
}

// HandleRangeRequest 处理范围请求（支持断点续传）
func HandleRangeRequest(w http.ResponseWriter, r *http.Request, musicFile *MusicFile) {
	// 解析Range头
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		http.Error(w, "Range header required", http.StatusBadRequest)
		return
	}

	// 这里实现范围请求处理逻辑
	// 暂时返回完整文件，后续可以优化为真正的流媒体支持
	http.Redirect(w, r, GetMusicFileURL(musicFile.StoragePath), http.StatusFound)
}

// DeleteMusicFile 删除音乐文件（从存储和数据库中删除）
func DeleteMusicFile(musicFileID string, userUUID string) error {
	// 首先获取文件信息
	httpClient := &http.Client{}
	getUrl := fmt.Sprintf("%s/rest/v1/music_files?id=eq.%s&user_id=eq.%s", os.Getenv("SUPABASE_URL"), musicFileID, userUUID)

	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return fmt.Errorf("创建获取请求失败: %v", err)
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("获取请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("音乐文件不存在，状态码: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result) == 0 {
		return fmt.Errorf("音乐文件不存在")
	}

	storagePath := getStringFromMapUpload(result[0], "storage_path", "")
	if storagePath == "" {
		return fmt.Errorf("无效的存储路径")
	}

	// 从存储中删除文件
	if err := deleteFromSupabaseStorage(storagePath); err != nil {
		return fmt.Errorf("删除存储文件失败: %v", err)
	}

	// 从数据库中删除记录
	delUrl := fmt.Sprintf("%s/rest/v1/music_files?id=eq.%s&user_id=eq.%s", os.Getenv("SUPABASE_URL"), musicFileID, userUUID)
	reqDel, err := http.NewRequest("DELETE", delUrl, nil)
	if err != nil {
		return fmt.Errorf("创建删除请求失败: %v", err)
	}

	reqDel.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	reqDel.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))

	respDel, err := httpClient.Do(reqDel)
	if err != nil {
		return fmt.Errorf("删除请求失败: %v", err)
	}
	defer respDel.Body.Close()

	if respDel.StatusCode != http.StatusNoContent && respDel.StatusCode != http.StatusOK {
		return fmt.Errorf("删除数据库记录失败，状态码: %d", respDel.StatusCode)
	}

	return nil
}

// 辅助函数：从map中安全获取字符串
func getStringFromMapUpload(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// 辅助函数：从map中安全获取int64
func getInt64FromMapUpload(m map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := m[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// 辅助函数：从map中安全获取int
func getIntFromMapUpload(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int64:
			return int(v)
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}