package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	supabase "github.com/supabase-community/supabase-go"
)

// Client 单例
var client *supabase.Client

// Init 初始化 Supabase 客户端（使用环境变量）
// 需要设置：SUPABASE_URL、SUPABASE_ANON_KEY
func Init() error {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_ANON_KEY")

	// 若缺失则尝试加载 .env 文件
	if url == "" || key == "" {
		_ = godotenv.Load(".env")
		if url == "" {
			url = os.Getenv("SUPABASE_URL")
		}
		if key == "" {
			key = os.Getenv("SUPABASE_ANON_KEY")
		}
	}

	if url == "" || key == "" {
		return fmt.Errorf("missing env SUPABASE_URL or SUPABASE_ANON_KEY (请设置环境变量或提供 .env 文件)")
	}

	var err error
	client, err = supabase.NewClient(url, key, nil)
	return err
}

// RegisterUser 使用 Auth 注册邮箱+密码用户
func RegisterUser(ctx context.Context, email, password string) error {
	if client == nil {
		if err := Init(); err != nil {
			return err
		}
	}
	_, err := client.Auth.Signup(gotruetypes.SignupRequest{
		Email:    email,
		Password: password,
	})
	return err
}

// LoginUser 使用 Auth 登录邮箱+密码用户
func LoginUser(ctx context.Context, email, password string) error {
	if client == nil {
		if err := Init(); err != nil {
			return err
		}
	}
	_, err := client.Auth.Token(gotruetypes.TokenRequest{
		Email:     email,
		Password:  password,
		GrantType: "password",
	})
	return err
}

// Configure 允许通过代码显式设置 Supabase URL 与 KEY
func Configure(url, key string) error {
	if url == "" || key == "" {
		return fmt.Errorf("supabase url/key required")
	}
	c, err := supabase.NewClient(url, key, nil)
	if err != nil {
		return err
	}
	client = c
	return nil
}

// GetClient 获取 Supabase 客户端
func GetClient() *supabase.Client {
	return client
}

// UpdateUserNickname 更新用户昵称
func UpdateUserNickname(userID int, nickname string) error {
	if client == nil {
		if err := Init(); err != nil {
			return err
		}
	}
	
	// 使用HTTP客户端直接调用Supabase REST API
	httpClient := &http.Client{}
	
	// 首先检查用户配置是否存在，如果不存在则创建，存在则更新
	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%d", os.Getenv("SUPABASE_URL"), userID)
	
	// 先检查是否存在
	reqCheck, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建检查请求失败: %v", err)
	}
	
	reqCheck.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	reqCheck.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	reqCheck.Header.Set("Content-Type", "application/json")
	
	respCheck, err := httpClient.Do(reqCheck)
	if err != nil {
		return fmt.Errorf("检查请求失败: %v", err)
	}
	defer respCheck.Body.Close()
	
	var existingProfiles []map[string]interface{}
	if respCheck.StatusCode == http.StatusOK {
		json.NewDecoder(respCheck.Body).Decode(&existingProfiles)
	}
	
	var req *http.Request
	if len(existingProfiles) > 0 {
		// 更新现有配置
		updateData := map[string]interface{}{
			"nickname": nickname,
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.Marshal(updateData)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %v", err)
		}
		
		req, err = http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建更新请求失败: %v", err)
		}
	} else {
		// 创建新配置 - 首先尝试从auth.users表获取真实邮箱
		email := getUserEmail(userID)
		if email == "" {
			email = fmt.Sprintf("user%d@example.com", userID) // 备用默认邮箱
		}
		
		insertData := map[string]interface{}{
			"user_id":   userID,
			"email":     email,
			"nickname":  nickname,
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.Marshal(insertData)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %v", err)
		}
		
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/rest/v1/user_profiles", os.Getenv("SUPABASE_URL")), bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建插入请求失败: %v", err)
		}
		req.Header.Set("Prefer", "return=representation")
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 处理不同的响应状态码
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		// 成功状态
		return nil
	case http.StatusNotFound:
		// 表不存在，尝试创建表后重试
		fmt.Printf("用户表不存在，尝试创建表后重试\n")
		if err := CreateUserProfilesTable(); err != nil {
			return fmt.Errorf("创建用户表失败: %v", err)
		}
		// 重试操作
		return UpdateUserNickname(userID, nickname)
	default:
		// 读取错误响应体获取更多信息
		body, _ := io.ReadAll(resp.Body)
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if message, ok := errorResp["message"].(string); ok {
				return fmt.Errorf("API返回错误: %s (状态码: %d)", message, resp.StatusCode)
			}
		}
		return fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
}

// UpdateUserNicknameByUUID 更新用户昵称（使用UUID格式）
func UpdateUserNicknameByUUID(userUUID, nickname string) error {
	if client == nil {
		if err := Init(); err != nil {
			return err
		}
	}
	
	// 使用HTTP客户端直接调用Supabase REST API
	httpClient := &http.Client{}
	
	// 首先检查用户配置是否存在，如果不存在则创建，存在则更新
	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%s", os.Getenv("SUPABASE_URL"), userUUID)
	
	// 先检查是否存在
	reqCheck, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建检查请求失败: %v", err)
	}
	
	reqCheck.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	reqCheck.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	reqCheck.Header.Set("Content-Type", "application/json")
	
	respCheck, err := httpClient.Do(reqCheck)
	if err != nil {
		return fmt.Errorf("检查请求失败: %v", err)
	}
	defer respCheck.Body.Close()
	
	var existingProfiles []map[string]interface{}
	if respCheck.StatusCode == http.StatusOK {
		json.NewDecoder(respCheck.Body).Decode(&existingProfiles)
	}
	
	var req *http.Request
	if len(existingProfiles) > 0 {
		// 更新现有配置
		updateData := map[string]interface{}{
			"nickname": nickname,
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.Marshal(updateData)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %v", err)
		}
		
		req, err = http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建更新请求失败: %v", err)
		}
	} else {
		// 创建新配置
		insertData := map[string]interface{}{
			"user_id":   userUUID,
			"email":     "user@example.com", // 使用默认邮箱
			"nickname":  nickname,
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.Marshal(insertData)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %v", err)
		}
		
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/rest/v1/user_profiles", os.Getenv("SUPABASE_URL")), bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建插入请求失败: %v", err)
		}
		req.Header.Set("Prefer", "return=representation")
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 处理不同的响应状态码
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		// 成功状态
		return nil
	case http.StatusNotFound:
		// 表不存在，尝试创建表后重试
		fmt.Printf("用户表不存在，尝试创建表后重试\n")
		if err := CreateUserProfilesTable(); err != nil {
			return fmt.Errorf("创建用户表失败: %v", err)
		}
		// 重试操作
		return UpdateUserNicknameByUUID(userUUID, nickname)
	default:
		// 读取错误响应体获取更多信息
		body, _ := io.ReadAll(resp.Body)
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if message, ok := errorResp["message"].(string); ok {
				return fmt.Errorf("API返回错误: %s (状态码: %d)", message, resp.StatusCode)
			}
		}
		return fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
}

// GetUserProfile 获取用户个人资料（兼容整数ID）
func GetUserProfile(userID int) (map[string]interface{}, error) {
	if client == nil {
		if err := Init(); err != nil {
			return nil, err
		}
	}
	
	// 使用HTTP客户端直接调用Supabase REST API获取用户信息
	httpClient := &http.Client{}
	
	// 首先尝试从user_profiles表获取昵称
	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%d&select=*", os.Getenv("SUPABASE_URL"), userID)
	
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
	
	var nickname string
	var email string
	
	if resp.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result) > 0 {
			// 从user_profiles表获取昵称
			nickname = getStringFromMap(result[0], "nickname", "")
		}
	}
	
	// 如果user_profiles表中没有昵称，尝试从auth.users表获取
	if nickname == "" {
		authUrl := fmt.Sprintf("%s/rest/v1/auth.users?id=eq.%d&select=*", os.Getenv("SUPABASE_URL"), userID)
		reqAuth, err := http.NewRequest("GET", authUrl, nil)
		if err == nil {
			reqAuth.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
			reqAuth.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
			reqAuth.Header.Set("Content-Type", "application/json")
			
			respAuth, err := httpClient.Do(reqAuth)
			if err == nil {
				defer respAuth.Body.Close()
				if respAuth.StatusCode == http.StatusOK {
					var authResult []map[string]interface{}
					if err := json.NewDecoder(respAuth.Body).Decode(&authResult); err == nil && len(authResult) > 0 {
						email = getStringFromMap(authResult[0], "email", "")
						
						// 尝试从raw_user_meta_data获取昵称
						if metaData, ok := authResult[0]["raw_user_meta_data"].(map[string]interface{}); ok {
							nickname = getStringFromMap(metaData, "nickname", "")
						}
					}
				}
			}
		}
	}
	
	// 如果都没有获取到昵称，使用默认值
	if nickname == "" {
		nickname = "用户"
	}
	if email == "" {
		email = "user" + fmt.Sprintf("%d", userID) + "@example.com"
	}
	
	userInfo := map[string]interface{}{
		"id":       userID,
		"nickname": nickname,
		"email":    email,
	}
	
	return userInfo, nil
}



// 辅助函数：从map中安全获取字符串
func getStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getUserEmail 获取用户邮箱
func getUserEmail(userID int) string {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/auth.users?id=eq.%d&select=email", os.Getenv("SUPABASE_URL"), userID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result) > 0 {
			if email, ok := result[0]["email"].(string); ok {
				return email
			}
		}
	}
	
	return ""
}

// GetUserProfileByUUID 通过UUID获取用户个人资料
func GetUserProfileByUUID(userUUID string) (map[string]interface{}, error) {
	if client == nil {
		if err := Init(); err != nil {
			return nil, err
		}
	}
	
	// 使用HTTP客户端直接调用Supabase REST API获取用户信息
	httpClient := &http.Client{}
	
	// 首先尝试从user_profiles表获取昵称（使用UUID查询）
	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%s&select=*", os.Getenv("SUPABASE_URL"), userUUID)
	
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
	
	var nickname string
	var email string
	
	if resp.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result) > 0 {
			// 从user_profiles表获取昵称
			nickname = getStringFromMap(result[0], "nickname", "")
			email = getStringFromMap(result[0], "email", "")
		}
	}
	
	// 如果user_profiles表中没有昵称，尝试从auth.users表获取
	if nickname == "" || email == "" {
		authUrl := fmt.Sprintf("%s/rest/v1/auth.users?id=eq.%s&select=*", os.Getenv("SUPABASE_URL"), userUUID)
		reqAuth, err := http.NewRequest("GET", authUrl, nil)
		if err == nil {
			reqAuth.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
			reqAuth.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_ANON_KEY"))
			reqAuth.Header.Set("Content-Type", "application/json")
			
			respAuth, err := httpClient.Do(reqAuth)
			if err == nil {
				defer respAuth.Body.Close()
				if respAuth.StatusCode == http.StatusOK {
					var authResult []map[string]interface{}
					if err := json.NewDecoder(respAuth.Body).Decode(&authResult); err == nil && len(authResult) > 0 {
						if email == "" {
							email = getStringFromMap(authResult[0], "email", "")
						}
						
						// 尝试从raw_user_meta_data获取昵称
						if nickname == "" {
							if metaData, ok := authResult[0]["raw_user_meta_data"].(map[string]interface{}); ok {
								nickname = getStringFromMap(metaData, "nickname", "")
							}
						}
					}
				}
			}
		}
	}
	
	// 如果都没有获取到昵称，使用默认值
	if nickname == "" {
		nickname = "用户"
	}
	if email == "" {
		email = "user@example.com"
	}
	
	userInfo := map[string]interface{}{
		"id":       userUUID,
		"nickname": nickname,
		"email":    email,
	}
	
	return userInfo, nil
}


