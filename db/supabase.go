package db

import (
	"context"
	"fmt"
	"os"

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
