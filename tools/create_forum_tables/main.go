package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// 检查环境变量
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_ANON_KEY")
	
	if url == "" || key == "" {
		fmt.Println("❌ 请设置 SUPABASE_URL 和 SUPABASE_ANON_KEY 环境变量")
		fmt.Println("   或者在项目根目录创建 .env 文件")
		return
	}
	
	fmt.Println("🚀 开始创建音乐论坛数据库表...")
	fmt.Printf("📊 项目URL: %s\n", url)
	
	// 创建论坛帖子表
	if err := createForumPostsTable(url, key); err != nil {
		fmt.Printf("❌ 创建论坛帖子表失败: %v\n", err)
		return
	}
	
	// 创建论坛回复表
	if err := createForumRepliesTable(url, key); err != nil {
		fmt.Printf("❌ 创建论坛回复表失败: %v\n", err)
		return
	}
	
	fmt.Println("✅ 音乐论坛数据库表创建完成！")
	fmt.Println("\n📋 创建的表结构：")
	fmt.Println("   - forum_posts (论坛帖子表)")
	fmt.Println("   - forum_replies (论坛回复表)")
	fmt.Println("\n🔐 行级安全策略已启用")
	fmt.Println("👥 用户权限策略已配置")
}

func createForumPostsTable(url, key string) error {
	fmt.Println("\n📝 创建论坛帖子表 (forum_posts)...")
	
	// SQL语句创建表
	sql := `
CREATE TABLE IF NOT EXISTS forum_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    user_nickname VARCHAR(100) NOT NULL,
    music_reference JSONB,
    tags VARCHAR[] DEFAULT '{}',
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT false,
    is_locked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 启用行级安全
ALTER TABLE forum_posts ENABLE ROW LEVEL SECURITY;

-- 创建策略：所有用户都可以查看帖子
CREATE POLICY "任何人都可以查看帖子" ON forum_posts
FOR SELECT USING (true);

-- 创建策略：认证用户可以创建帖子
CREATE POLICY "认证用户可以创建帖子" ON forum_posts
FOR INSERT WITH CHECK (auth.uid() = user_id);

-- 创建策略：用户只能编辑自己的帖子
CREATE POLICY "用户只能编辑自己的帖子" ON forum_posts
FOR UPDATE USING (auth.uid() = user_id);

-- 创建策略：用户只能删除自己的帖子
CREATE POLICY "用户只能删除自己的帖子" ON forum_posts
FOR DELETE USING (auth.uid() = user_id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_forum_posts_created_at ON forum_posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_forum_posts_user_id ON forum_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_tags ON forum_posts USING GIN(tags);
`

	return executeSQL(url, key, sql)
}

func createForumRepliesTable(url, key string) error {
	fmt.Println("\n💬 创建论坛回复表 (forum_replies)...")
	
	// SQL语句创建表
	sql := `
CREATE TABLE IF NOT EXISTS forum_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    user_nickname VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES forum_replies(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 启用行级安全
ALTER TABLE forum_replies ENABLE ROW LEVEL SECURITY;

-- 创建策略：所有用户都可以查看回复
CREATE POLICY "任何人都可以查看回复" ON forum_replies
FOR SELECT USING (true);

-- 创建策略：认证用户可以创建回复
CREATE POLICY "认证用户可以创建回复" ON forum_replies
FOR INSERT WITH CHECK (auth.uid() = user_id);

-- 创建策略：用户只能编辑自己的回复
CREATE POLICY "用户只能编辑自己的回复" ON forum_replies
FOR UPDATE USING (auth.uid() = user_id);

-- 创建策略：用户只能删除自己的回复
CREATE POLICY "用户只能删除自己的回复" ON forum_replies
FOR DELETE USING (auth.uid() = user_id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_forum_replies_post_id ON forum_replies(post_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_user_id ON forum_replies(user_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_parent_id ON forum_replies(parent_id);
CREATE INDEX IF NOT EXISTS idx_forum_replies_created_at ON forum_replies(created_at);
`

	return executeSQL(url, key, sql)
}

func executeSQL(url, key, sql string) error {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	
	// 准备请求
	reqBody := map[string]interface{}{
		"query": sql,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化SQL失败: %v", err)
	}
	
	req, err := http.NewRequest("POST", url+"/rest/v1/rpc/exec_sql", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")
	
	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("执行SQL失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if message, ok := errorResp["message"].(string); ok {
				return fmt.Errorf("SQL执行错误: %s", message)
			}
		}
		return fmt.Errorf("HTTP错误状态码: %d", resp.StatusCode)
	}
	
	return nil
}