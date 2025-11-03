# Coze智能体配置指南

## 配置步骤

### 1. 获取Coze API密钥
1. 访问 [Coze开放平台](https://www.coze.cn/open)
2. 注册或登录你的账号
3. 进入"开发者中心" -> "API密钥管理"
4. 创建一个新的API密钥

### 2. 配置API密钥
编辑 `config/coze_config.go` 文件，将 `your_coze_api_key_here` 替换为你的实际API密钥：

```go
var CozeConfig = struct {
	APIBaseURL string
	APIKey     string
	BotID      string
}{
	APIBaseURL: "https://api.coze.cn/v1",
	APIKey:     "你的实际API密钥", // 替换这里
	BotID:      "7567959568312369215",
}
```

### 3. 重启应用
配置完成后，重启你的音乐播放器应用：

```bash
go run main.go
```

## 测试AI助手

配置完成后，你可以：

1. 打开首页 (http://localhost:8080)
2. 在AI助手输入框中输入问题
3. 点击发送或按Enter键
4. AI助手将使用Coze智能体进行真实对话

## 故障排除

### 如果AI助手无法正常工作：

1. **检查API密钥**：确保API密钥正确且未过期
2. **检查网络连接**：确保可以访问Coze API
3. **查看控制台日志**：检查是否有错误信息
4. **备用回复**：如果Coze API不可用，系统会自动使用备用回复

### 备用回复功能
当Coze API不可用时，AI助手会自动切换到备用回复模式，提供基本的音乐相关回答。

## 支持的对话类型

AI助手可以处理以下类型的对话：
- 歌曲推荐
- 歌词查询
- 歌手信息
- 音乐知识
- 音乐制作技巧
- 歌曲分析
- 其他音乐相关问题

配置完成后，你的AI助手就可以与Coze智能体进行真实的智能对话了！