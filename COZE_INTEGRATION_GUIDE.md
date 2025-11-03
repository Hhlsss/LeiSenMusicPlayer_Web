# Coze智能体集成指南

## 概述

本文档指导你如何将Coze智能体集成到音乐播放器网站的AI助手功能中。

## 已完成的功能

1. **后端API接口** (`controller/ai.go`)
   - 创建了 `/api/ai/chat` 端点
   - 集成了Coze智能体API调用逻辑
   - 包含错误处理和备用回复机制

2. **前端集成** (`web/index.html`)
   - 更新了AI助手界面，从模拟回复改为真实的Coze智能体回复
   - 添加了加载状态和错误处理
   - 优化了用户体验

3. **样式优化** (`web/styles.css`)
   - 添加了加载动画效果
   - 优化了消息显示效果

## 需要你配置的信息

### 1. 获取Coze API密钥

1. 访问 [Coze开放平台](https://www.coze.cn/open)
2. 登录你的账号
3. 进入"账号设置" → "API密钥"
4. 创建新的API密钥

### 2. 配置API信息

打开 `controller/ai.go` 文件，找到以下配置部分：

```go
// Coze API配置
const (
    CozeAPIBaseURL = "https://api.coze.cn/v1"
    // 这里需要你提供实际的API密钥
    CozeAPIKey = "your_coze_api_key_here"  // ← 替换为你的API密钥
)
```

将 `your_coze_api_key_here` 替换为你在Coze平台获取的实际API密钥。

### 3. 智能体ID配置

在同一个文件中，确保智能体ID正确：

```go
requestBody := CozeChatRequest{
    BotID:   "7567959568312369215", // 你的智能体ID - 这个已经正确配置
    Message: message,
    Stream:  false,
}
```

## API端点信息

根据Coze官方文档，标准的API端点格式为：

- **基础URL**: `https://api.coze.cn/v1`
- **聊天端点**: `/chat/completions`
- **认证方式**: Bearer Token (API密钥)

## 测试集成

1. **启动服务器**:
   ```bash
   go run main.go
   ```

2. **访问首页**: http://localhost:8080

3. **测试AI助手**: 在首页的AI助手区域输入消息进行测试

## 故障排除

### 如果Coze API不可用

系统会自动切换到备用回复模式，提供基本的音乐相关回复。

### 常见错误

1. **401 Unauthorized**: API密钥错误或过期
2. **404 Not Found**: API端点URL错误
3. **429 Too Many Requests**: 请求频率超限

### 调试方法

1. 检查浏览器开发者工具的网络面板
2. 查看服务器控制台日志
3. 验证API密钥是否正确配置

## 下一步

1. 获取并配置你的Coze API密钥
2. 测试AI助手功能
3. 根据实际需求调整回复逻辑
4. 考虑添加对话历史记录功能

## 安全建议

- 不要将API密钥提交到版本控制系统
- 考虑使用环境变量存储敏感信息
- 在生产环境中启用HTTPS

## 技术支持

如果遇到问题，请检查：
- Coze API密钥是否有效
- 网络连接是否正常
- 服务器日志中的错误信息