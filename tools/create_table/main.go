package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	fmt.Println("正在尝试创建 user_profiles 表...")
	
	// 首先检查表是否已存在
	if checkTableExists() {
		fmt.Println("✓ user_profiles 表已存在")
		return
	}
	
	// 创建表
	if err := createTable(); err != nil {
		fmt.Printf("✗ 创建表失败: %v\n", err)
		fmt.Println("\n请手动在 Supabase 仪表板中创建表：")
		fmt.Println("1. 登录 https://supabase.com/dashboard")
		fmt.Println("2. 选择您的项目")
		fmt.Println("3. 进入 Table Editor")
		fmt.Println("4. 点击 Create a new table")
		fmt.Println("5. 设置表名为: user_profiles")
		fmt.Println("6. 添加以下字段：")
		fmt.Println("   - user_id (integer, Primary Key)")
		fmt.Println("   - email (text, Not Null)")
		fmt.Println("   - nickname (text, Not Null)")
		fmt.Println("   - created_at (timestamptz, Default: now())")
		fmt.Println("   - updated_at (timestamptz, Default: now())")
		return
	}
	
	fmt.Println("✓ user_profiles 表创建成功")
}

func checkTableExists() bool {
	httpClient := &http.Client{}
	
	url := fmt.Sprintf("%s/rest/v1/user_profiles?limit=1", getSupabaseURL())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("apikey", getSupabaseAnonKey())
	req.Header.Set("Authorization", "Bearer "+getSupabaseAnonKey())
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

func createTable() error {
	httpClient := &http.Client{}
	
	// 使用 SQL API 创建表
	sql := `CREATE TABLE user_profiles (
		user_id INTEGER PRIMARY KEY,
		email VARCHAR NOT NULL,
		nickname VARCHAR NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`
	
	requestData := map[string]interface{}{
		"query": sql,
	}
	
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化SQL语句失败: %v", err)
	}
	
	url := fmt.Sprintf("%s/rest/v1/", getSupabaseURL())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	
	req.Header.Set("apikey", getSupabaseAnonKey())
	req.Header.Set("Authorization", "Bearer "+getSupabaseAnonKey())
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