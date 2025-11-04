package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Coze API配置
const (
	CozeAPIBaseURL = "https://api.coze.cn"
	CozeAPIKey     = "pat_1mZRTEMafM2v3ZjLZa2ioCoNRBp86Dh3bKl3X4667qtYgB0FLbOVW90fdkppYsaq"
	BotID          = "7567959568312369215"
	SpaceID        = "7559750439882375218"
)

// 调试信息
var (
	DebugMode = true
)

// CozeChatRequest Coze API请求结构
type CozeChatRequest struct {
	ConversationID string `json:"conversation_id"`
	BotID          string `json:"bot_id"`
	User           string `json:"user"`
	Query          string `json:"query"`
	Stream         bool   `json:"stream"`
}

// CozeChatResponse Coze API响应结构
type CozeChatResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Messages []struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"messages"`
	} `json:"data"`
}

// OpenAI格式请求结构
type OpenAIChatRequest struct {
	Model    string                  `json:"model"`
	Messages []OpenAIChatMessage     `json:"messages"`
	Stream   bool                   `json:"stream"`
}

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI格式响应结构
type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// AI聊天请求结构
type AIChatRequest struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// AI聊天响应结构
type AIChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// HandleAIChat AI聊天接口
func HandleAIChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAIError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		sendAIError(w, "消息内容不能为空", http.StatusBadRequest)
		return
	}

	// 调用Coze API
	response, err := callCozeAPI(req.Message, req.UserID)
	if err != nil {
		log.Printf("Coze API调用失败: %v", err)
		
		// 在调试模式下记录详细错误，但仍然使用备用回复
		if DebugMode {
			log.Printf("调试模式：Coze API调用失败，使用备用回复")
		}
		
		// 尝试备用回复
		fallbackResponse := getFallbackResponse(req.Message)
		sendAISuccess(w, fallbackResponse)
		return
	}

	sendAISuccess(w, response)
}

// 调用Coze API
func callCozeAPI(message, userID string) (string, error) {
	// 尝试多个API端点
	endpoints := []struct {
		name   string
		url    string
		format string
	}{
		{
			name:   "Coze API v2 with bot_id",
			url:    CozeAPIBaseURL + "/open_api/v2/chat",
			format: "coze_v2_bot",
		},
		{
			name:   "Coze API v2 with bot_id in URL",
			url:    CozeAPIBaseURL + "/open_api/v2/chat?bot_id=" + BotID,
			format: "coze_v2_url",
		},
	}

	for _, endpoint := range endpoints {
		log.Printf("尝试API端点: %s - URL: %s", endpoint.name, endpoint.url)
		
		var requestBody []byte
		var err error
		
		switch endpoint.format {
		case "coze_v2_bot":
			// 使用bot_id作为查询参数的格式
			requestBody, err = json.Marshal(map[string]interface{}{
				"conversation_id": fmt.Sprintf("music_assistant_%s_%d", userID, time.Now().Unix()),
				"bot_id":          BotID,
				"user":            "music_web_user",
				"query":           message,
				"stream":          false,
			})
		case "coze_v2_url":
			// 使用bot_id在URL中的格式
			requestBody, err = json.Marshal(map[string]interface{}{
				"conversation_id": fmt.Sprintf("music_assistant_%s_%d", userID, time.Now().Unix()),
				"user":            "music_web_user",
				"query":           message,
				"stream":          false,
			})
		}
		
		if err != nil {
			log.Printf("序列化请求体失败: %v", err)
			continue
		}

		if DebugMode {
			log.Printf("请求体: %s", string(requestBody))
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("POST", endpoint.url, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Printf("创建请求失败: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+CozeAPIKey)

		if DebugMode {
			log.Printf("请求头: Authorization: Bearer %s", CozeAPIKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("API请求失败: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("API响应错误: %d - %s", resp.StatusCode, string(body))
			continue
		}

		// 解析响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取响应体失败: %v", err)
			continue
		}

		if DebugMode {
			log.Printf("API响应: %s", string(body))
		}

		// 根据端点格式解析响应
		switch endpoint.format {
		case "coze_v2_bot":
			// 解析新的Coze API v2响应格式
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				log.Printf("解析Coze v2 bot响应失败: %v", err)
				continue
			}
			
			if DebugMode {
				log.Printf("Coze v2 bot 完整响应: %+v", response)
			}
			
			if code, ok := response["code"].(float64); ok && code == 0 {
				// 新的Coze API v2格式：messages直接在根级别
				if messages, ok := response["messages"].([]interface{}); ok && len(messages) > 0 {
					// 查找第一个类型为"answer"的消息
					for _, msg := range messages {
						if message, ok := msg.(map[string]interface{}); ok {
							if msgType, ok := message["type"].(string); ok && msgType == "answer" {
								if content, ok := message["content"].(string); ok {
									log.Printf("Coze API v2 bot 调用成功")
									return content, nil
								} else {
									log.Printf("Coze v2 bot: answer消息内容不是字符串类型")
								}
							}
						}
					}
					// 如果没有找到answer类型的消息，使用第一条消息
					if firstMsg, ok := messages[0].(map[string]interface{}); ok {
						if content, ok := firstMsg["content"].(string); ok {
							log.Printf("Coze API v2 bot 调用成功（使用第一条消息）")
							return content, nil
						} else {
							log.Printf("Coze v2 bot: 第一条消息内容不是字符串类型")
						}
					} else {
						log.Printf("Coze v2 bot: 第一条消息格式错误")
					}
				} else {
					log.Printf("Coze v2 bot: messages字段不存在或不是数组")
				}
			} else {
				log.Printf("Coze API v2 bot 响应异常，code: %v, message: %v", response["code"], response["message"])
			}
		case "coze_v2_url":
			// 解析bot_id在URL中的响应格式
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				log.Printf("解析Coze v2 URL响应失败: %v", err)
				continue
			}
			
			if DebugMode {
				log.Printf("Coze v2 URL 完整响应: %+v", response)
			}
			
			if code, ok := response["code"].(float64); ok && code == 0 {
				// 新的Coze API v2格式：messages直接在根级别
				if messages, ok := response["messages"].([]interface{}); ok && len(messages) > 0 {
					// 查找第一个类型为"answer"的消息
					for _, msg := range messages {
						if message, ok := msg.(map[string]interface{}); ok {
							if msgType, ok := message["type"].(string); ok && msgType == "answer" {
								if content, ok := message["content"].(string); ok {
									log.Printf("Coze API v2 URL 调用成功")
									return content, nil
								} else {
									log.Printf("Coze v2 URL: answer消息内容不是字符串类型")
								}
							}
						}
					}
					// 如果没有找到answer类型的消息，使用第一条消息
					if firstMsg, ok := messages[0].(map[string]interface{}); ok {
						if content, ok := firstMsg["content"].(string); ok {
							log.Printf("Coze API v2 URL 调用成功（使用第一条消息）")
							return content, nil
						} else {
							log.Printf("Coze v2 URL: 第一条消息内容不是字符串类型")
						}
					} else {
						log.Printf("Coze v2 URL: 第一条消息格式错误")
					}
				} else {
					log.Printf("Coze v2 URL: messages字段不存在或不是数组")
				}
			} else {
				log.Printf("Coze API v2 URL 响应异常，code: %v, message: %v", response["code"], response["message"])
			}
		}
	}

	return "", fmt.Errorf("所有API端点都失败，请检查API密钥和智能体配置")
}

// 备用回复（当API不可用时使用）
func getFallbackResponse(message string) string {
	// 简单的关键词匹配回复
	keywords := map[string]string{
		"推荐":    "根据您的听歌历史，我为您推荐一些相似风格的歌曲。",
		"歌单":    "我可以帮您创建和管理歌单，请告诉我您喜欢的音乐类型。",
		"歌手":    "请告诉我您想了解的歌手名字，我可以为您介绍相关信息。",
		"专辑":    "我可以为您推荐热门专辑或根据您的喜好推荐专辑。",
		"播放":    "您可以直接点击歌曲进行播放，或告诉我您想听什么类型的音乐。",
		"下载":    "目前支持在线播放功能，下载功能正在开发中。",
		"帮助":    "我是您的智能音乐助手，可以帮您：推荐歌曲、管理歌单、解答音乐问题等。",
	}

	for keyword, response := range keywords {
		if contains(message, keyword) {
			return response
		}
	}

	// 默认回复
	return "我是您的智能音乐助手，可以帮您推荐歌曲、管理歌单、解答音乐问题等。请告诉我您需要什么帮助？"
}

// 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
			s[len(s)-len(substr):] == substr)))
}

// 发送成功响应
func sendAISuccess(w http.ResponseWriter, message string) {
	response := AIChatResponse{
		Success: true,
		Message: message,
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 发送错误响应
func sendAIError(w http.ResponseWriter, errorMsg string, statusCode int) {
	response := AIChatResponse{
		Success: false,
		Message: "",
		Error:   errorMsg,
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// HandleAIChatTest AI聊天测试接口
func HandleAIChatTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testMessage := "你好，请介绍一下你自己"
	response, err := callCozeAPI(testMessage, "test_user")
	
	result := map[string]interface{}{
		"test_message": testMessage,
		"success":      err == nil,
		"response":     response,
		"error":        "",
	}
	
	if err != nil {
		result["error"] = err.Error()
	}
	
	json.NewEncoder(w).Encode(result)
}