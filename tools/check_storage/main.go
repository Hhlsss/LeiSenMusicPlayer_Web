package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	// 检查环境变量
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		fmt.Println("错误: 请设置 SUPABASE_URL 和 SUPABASE_ANON_KEY 环境变量")
		os.Exit(1)
	}

	fmt.Printf("正在检查 Supabase 存储桶...\n")
	fmt.Printf("项目URL: %s\n", supabaseURL)
	fmt.Printf("存储桶名称: music\n\n")

	// 检查存储桶是否存在
	bucketExists, err := checkBucketExists(supabaseURL, supabaseKey)
	if err != nil {
		fmt.Printf("检查存储桶失败: %v\n", err)
		os.Exit(1)
	}

	if !bucketExists {
		fmt.Println("❌ 存储桶 'music' 不存在")
		fmt.Println("请按照 SUPABASE_STORAGE_SETUP.md 中的说明创建存储桶")
		os.Exit(1)
	}

	fmt.Println("✅ 存储桶 'music' 存在")

	// 列出存储桶中的文件
	files, err := listBucketFiles(supabaseURL, supabaseKey)
	if err != nil {
		fmt.Printf("列出文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("\n📁 存储桶中没有文件")
	} else {
		fmt.Printf("\n📁 存储桶中有 %d 个文件:\n", len(files))
		for i, file := range files {
			fmt.Printf("%d. %s (大小: %s)\n", i+1, file.Name, formatFileSize(file.Size))
		}
	}

	// 检查是否有音乐文件
	musicFiles := filterMusicFiles(files)
	if len(musicFiles) == 0 {
		fmt.Println("\n🎵 没有找到音乐文件")
	} else {
		fmt.Printf("\n🎵 找到 %d 个音乐文件:\n", len(musicFiles))
		for i, file := range musicFiles {
			fmt.Printf("%d. %s (类型: %s)\n", i+1, file.Name, getFileType(file.Name))
		}
	}
}

// 检查存储桶是否存在
func checkBucketExists(supabaseURL, supabaseKey string) (bool, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/storage/v1/bucket/music", supabaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// 列出存储桶中的文件
func listBucketFiles(supabaseURL, supabaseKey string) ([]FileInfo, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/storage/v1/object/list/music", supabaseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []FileInfo `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// 文件信息结构
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// 过滤音乐文件
func filterMusicFiles(files []FileInfo) []FileInfo {
	var musicFiles []FileInfo
	musicExtensions := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac", ".wma"}

	for _, file := range files {
		for _, ext := range musicExtensions {
			if strings.HasSuffix(strings.ToLower(file.Name), ext) {
				musicFiles = append(musicFiles, file)
				break
			}
		}
	}

	return musicFiles
}

// 获取文件类型
func getFileType(filename string) string {
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	switch ext {
	case ".mp3":
		return "MP3"
	case ".wav":
		return "WAV"
	case ".flac":
		return "FLAC"
	case ".ogg":
		return "OGG"
	case ".m4a":
		return "M4A"
	case ".aac":
		return "AAC"
	case ".wma":
		return "WMA"
	default:
		return "未知"
	}
}

// 格式化文件大小
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}