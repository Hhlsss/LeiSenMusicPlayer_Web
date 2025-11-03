package config

// CozeConfig 存储Coze API配置
var CozeConfig = struct {
	APIBaseURL string
	APIKey     string
	BotID      string
}{
	APIBaseURL: "https://api.coze.cn/v1",
	APIKey:     "your_coze_api_key_here", // 请替换为你的实际API密钥
	BotID:      "7567959568312369215",    // 智能体ID
}