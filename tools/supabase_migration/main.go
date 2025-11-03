package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// SQLExecutionRequest 执行SQL的请求结构
type SQLExecutionRequest struct {
	Query string `json:"query"`
}

func main() {
	fmt.Println("🚀 开始执行Supabase数据库迁移（方案1：整数ID → UUID）")
	
	// 配置Supabase连接
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseAccessToken := os.Getenv("SUPABASE_ACCESS_TOKEN")
	
	if supabaseURL == "" || supabaseAnonKey == "" || supabaseAccessToken == "" {
		fmt.Println("❌ 缺少必要的环境变量")
		fmt.Println("请设置: SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_ACCESS_TOKEN")
		os.Exit(1)
	}
	
	fmt.Printf("📊 连接到Supabase项目: %s\n", supabaseURL)
	
	// 读取SQL脚本
	sqlScript, err := os.ReadFile("supabase_migration_script.sql")
	if err != nil {
		fmt.Printf("❌ 无法读取SQL脚本文件: %v\n", err)
		os.Exit(1)
	}
	
	// 分割SQL语句
	sqlStatements := strings.Split(string(sqlScript), ";")
	
	// 执行每个SQL语句
	for i, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		fmt.Printf("\n📝 执行SQL语句 %d/%d:\n", i+1, len(sqlStatements))
		fmt.Printf("   %s\n", truncateString(stmt, 100))
		
		// 执行SQL
		if err := executeSQL(supabaseURL, supabaseAnonKey, supabaseAccessToken, stmt); err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			
			// 如果是SELECT语句，继续执行
			if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(stmt)), "SELECT") {
				fmt.Println("⚠️  SELECT语句可能返回空结果，继续执行...")
				continue
			}
			
			fmt.Println("❌ 迁移失败，请检查错误信息")
			os.Exit(1)
		}
		
		fmt.Println("✅ 执行成功")
		
		// 添加延迟避免请求过快
		time.Sleep(500 * time.Millisecond)
	}
	
	fmt.Println("\n🎉 数据库迁移完成！")
	fmt.Println("✅ 表结构已成功修改为UUID格式")
	fmt.Println("✅ 行级安全策略已配置")
	fmt.Println("✅ 索引已创建")
	fmt.Println("\n📋 下一步：请更新Go代码中的ID处理逻辑")
}

// executeSQL 执行单个SQL语句
func executeSQL(url, anonKey, accessToken, sql string) error {
	// 构建请求体
	requestBody := SQLExecutionRequest{
		Query: sql,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequest("POST", url+"/rest/v1/rpc/execute_sql", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Prefer", "return=representation")
	
	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if message, ok := errorResp["message"].(string); ok {
				return fmt.Errorf("API返回错误: %s (状态码: %d)", message, resp.StatusCode)
			}
		}
		return fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	// 打印响应结果（如果有）
	if len(body) > 0 && string(body) != "null" && string(body) != "[]" {
		fmt.Printf("📊 响应结果: %s\n", string(body))
	}
	
	return nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}