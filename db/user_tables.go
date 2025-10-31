package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// CreateUserProfilesTable 检查用户个人资料表是否存在，如果不存在则提供创建指导
func CreateUserProfilesTable() error {
	httpClient := &http.Client{}
	
	// 检查表是否存在
	checkUrl := fmt.Sprintf("%s/rest/v1/user_profiles?limit=1", os.Getenv("SUPABASE_URL"))
	reqCheck, err := http.NewRequest("GET", checkUrl, nil)
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
	
	// 如果表已存在，直接返回成功
	if respCheck.StatusCode == http.StatusOK {
		fmt.Println("用户个人资料表已存在")
		return nil
	}
	
	// 表不存在，提供详细的创建指导
	return fmt.Errorf(`user_profiles 表不存在。请在 Supabase 仪表板中手动创建该表。

创建步骤：
1. 登录 Supabase 控制台: https://supabase.com/dashboard
2. 选择您的项目: %s
3. 进入 "Table Editor"
4. 点击 "Create a new table"
5. 设置表名为: user_profiles
6. 添加以下字段：
   - user_id (integer, Primary Key)
   - email (text, Not Null)
   - nickname (text, Not Null)
   - created_at (timestamptz, Default: now())
   - updated_at (timestamptz, Default: now())

或者使用 SQL 在 Supabase SQL Editor 中执行：
CREATE TABLE user_profiles (
    user_id INTEGER PRIMARY KEY,
    email VARCHAR NOT NULL,
    nickname VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`, os.Getenv("SUPABASE_URL"))
}

// CreateOrUpdateUserProfile 创建或更新用户个人资料
func CreateOrUpdateUserProfile(userID int, email, nickname string) error {
	if client == nil {
		if err := Init(); err != nil {
			return err
		}
	}

	// 首先尝试创建用户表
	if err := CreateUserProfilesTable(); err != nil {
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	// 使用HTTP客户端直接调用Supabase REST API
	httpClient := &http.Client{}

	// 检查用户配置是否存在
	url := fmt.Sprintf("%s/rest/v1/user_profiles?user_id=eq.%d", os.Getenv("SUPABASE_URL"), userID)
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
			"nickname":  nickname,
			"email":     email,
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
		return CreateOrUpdateUserProfile(userID, email, nickname)
	default:
		// 读取错误响应体获取更多信息
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if message, ok := errorResp["message"].(string); ok {
				return fmt.Errorf("API返回错误: %s (状态码: %d)", message, resp.StatusCode)
			}
		}
		return fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
}

// GetUserProfileByEmail 根据邮箱获取用户个人资料
func GetUserProfileByEmail(email string) (map[string]interface{}, error) {
	if client == nil {
		if err := Init(); err != nil {
			return nil, err
		}
	}

	// 使用HTTP客户端直接调用Supabase REST API获取用户信息
	httpClient := &http.Client{}

	// 从user_profiles表获取用户信息
	url := fmt.Sprintf("%s/rest/v1/user_profiles?email=eq.%s&select=*", os.Getenv("SUPABASE_URL"), email)

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

	if resp.StatusCode == http.StatusOK {
		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result) > 0 {
			return result[0], nil
		}
	}

	return nil, fmt.Errorf("用户不存在")
}