package service

import (
	"context"

	"MusicPlayerWeb/db"
)

// RegisterUser 业务逻辑：注册用户（邮箱/账号 + 密码）
// 这里将 account 视为邮箱字段
func RegisterUser(ctx context.Context, account, password string) error {
	return db.RegisterUser(ctx, account, password)
}

// LoginUser 业务逻辑：登录用户
func LoginUser(ctx context.Context, account, password string) error {
	return db.LoginUser(ctx, account, password)
}
