package main

import (
	"fmt"
	"net/http"
	"os"
)

// 从环境变量获取 Supabase 配置
func getSupabaseURL() string {
	if url := os.Getenv("SUPABASE_URL"); url != "" {
		return url
	}
	return "https://gblnpzstdjnvclijjpbk.supabase.co"
}

func getSupabaseAnonKey() string {
	if key := os.Getenv("SUPABASE_ANON_KEY"); key != "" {
		return key
	}
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdibG5wenN0ZGpudmNsaWpqcGJrIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0NjgzNjEsImV4cCI6MjA3NjA0NDM2MX0.YaMwiKVVUgToza4vMtBjHAgvE28fdWTqtNHHtI9peFU"
}

func main() {
	fmt.Println("正在检查 user_profiles 表是否存在...")
	
	httpClient := &http.Client{}
	
	url := fmt.Sprintf("%s/rest/v1/user_profiles?limit=1", SUPABASE_URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("✗ 创建请求失败: %v\n", err)
		return
	}
	
	req.Header.Set("apikey", SUPABASE_ANON_KEY)
	req.Header.Set("Authorization", "Bearer "+SUPABASE_ANON_KEY)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("✗ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ user_profiles 表已存在")
		fmt.Println("\n表已存在，昵称修改功能应该可以正常工作了！")
		fmt.Println("\n如果仍有问题，请检查：")
		fmt.Println("1. 应用是否重新启动")
		fmt.Println("2. 用户是否已登录")
		fmt.Println("3. 网络连接是否正常")
	} else {
		fmt.Printf("✗ user_profiles 表不存在 (状态码: %d)\n", resp.StatusCode)
		fmt.Println("\n请按照以下步骤创建表：")
		fmt.Println("1. 登录 https://supabase.com/dashboard")
		fmt.Println("2. 选择项目: gblnpzstdjnvclijjpbk")
		fmt.Println("3. 进入 Table Editor")
		fmt.Println("4. 点击 Create a new table")
		fmt.Println("5. 设置表名为: user_profiles")
		fmt.Println("6. 添加以下字段：")
		fmt.Println("   - user_id (integer, Primary Key)")
		fmt.Println("   - email (text, Not Null)")
		fmt.Println("   - nickname (text, Not Null)")
		fmt.Println("   - created_at (timestamptz, Default: now())")
		fmt.Println("   - updated_at (timestamptz, Default: now())")
		fmt.Println("\n详细指南请查看 SUPABASE_TABLE_SETUP.md 文件")
	}
}