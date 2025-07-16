package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"net/http"
	"time"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Finished  bool   `json:"finished"`
}

var globalConfig *Config
var openaiClient *openai.Client

func main() {
	// 加载配置文件
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证必要的配置
	if config.ChatModels.Configs.APIKey == "" {
		log.Fatal("错误：API密钥不能为空，请在config.yaml中设置或使用环境变量OPENAI_API_KEY")
	}

	globalConfig = config

	fmt.Printf("使用配置:\n")
	fmt.Printf("- 平台: %s\n", config.ChatModels.Platform)
	fmt.Printf("- 代理URL: %s\n", config.ChatModels.Configs.ProxyURL)
	fmt.Printf("- 默认模型: %s\n", config.ChatModels.Configs.DefaultModel)
	fmt.Printf("- 最大Token: %d\n", config.Settings.MaxTokens)
	fmt.Printf("- 服务端口: %d\n", config.Settings.Port)
	fmt.Println("")

	// 创建自定义配置的OpenAI客户端
	clientConfig := openai.DefaultConfig(config.ChatModels.Configs.APIKey)
	clientConfig.BaseURL = config.ChatModels.Configs.ProxyURL
	openaiClient = openai.NewClientWithConfig(clientConfig)

	// 设置路由
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/chat", handleChat)
	http.HandleFunc("/stream", handleSSE)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Settings.Port)
	fmt.Printf("服务器启动在 http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// 提供主页面
func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

// 处理聊天请求
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// 处理SSE连接
func handleSSE(w http.ResponseWriter, r *http.Request) {
	// 设置SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 获取查询参数中的消息
	message := r.URL.Query().Get("message")
	if message == "" {
		message = "Who is Isaac Newton?"
	}

	ctx := context.Background()

	chatReq := openai.ChatCompletionRequest{
		Model:     globalConfig.ChatModels.Configs.DefaultModel,
		MaxTokens: globalConfig.Settings.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
		Stream: globalConfig.Settings.Stream,
	}

	stream, err := openaiClient.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", formatSSEError(err))
		return
	}
	defer stream.Close()

	// 发送开始事件
	fmt.Fprintf(w, "data: %s\n\n", formatSSEResponse("", false))
	w.(http.Flusher).Flush()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// 发送结束事件
			fmt.Fprintf(w, "data: %s\n\n", formatSSEResponse("", true))
			w.(http.Flusher).Flush()
			return
		}

		if err != nil {
			fmt.Fprintf(w, "data: %s\n\n", formatSSEError(err))
			w.(http.Flusher).Flush()
			return
		}

		// 发送内容
		content := response.Choices[0].Delta.Content
		if content != "" {
			fmt.Fprintf(w, "data: %s\n\n", formatSSEResponse(content, false))
			w.(http.Flusher).Flush()
		}
	}
}

// 格式化SSE响应
func formatSSEResponse(content string, finished bool) string {
	response := ChatResponse{
		Content:   content,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Finished:  finished,
	}
	data, _ := json.Marshal(response)
	return string(data)
}

// 格式化SSE错误
func formatSSEError(err error) string {
	errorResponse := map[string]interface{}{
		"error":     err.Error(),
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"finished":  true,
	}
	data, _ := json.Marshal(errorResponse)
	return string(data)
}

// HTML页面
const indexHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenAI SSE Chat Demo</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .chat-box {
            border: 1px solid #ddd;
            border-radius: 5px;
            height: 400px;
            overflow-y: auto;
            padding: 10px;
            margin-bottom: 20px;
            background-color: #fafafa;
        }
        .message {
            margin-bottom: 10px;
            padding: 8px;
            border-radius: 5px;
        }
        .user-message {
            background-color: #007bff;
            color: white;
            text-align: right;
        }
        .ai-message {
            background-color: #e9ecef;
            color: #333;
        }
        .input-group {
            display: flex;
            gap: 10px;
        }
        input[type="text"] {
            flex: 1;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
        }
        button {
            padding: 10px 20px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover {
            background-color: #0056b3;
        }
        button:disabled {
            background-color: #6c757d;
            cursor: not-allowed;
        }
        .timestamp {
            font-size: 12px;
            color: #666;
            margin-top: 5px;
        }
        .status {
            text-align: center;
            color: #666;
            font-style: italic;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>OpenAI SSE Chat Demo</h1>
        <div id="chatBox" class="chat-box"></div>
        <div class="input-group">
            <input type="text" id="messageInput" placeholder="输入您的消息..." onkeypress="handleKeyPress(event)">
            <button id="sendButton" onclick="sendMessage()">发送</button>
        </div>
        <div id="status" class="status"></div>
    </div>

    <script>
        let eventSource = null;
        let isConnected = false;

        function addMessage(content, isUser, timestamp) {
            const chatBox = document.getElementById('chatBox');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message ' + (isUser ? 'user-message' : 'ai-message');
            
            const contentDiv = document.createElement('div');
            contentDiv.textContent = content;
            messageDiv.appendChild(contentDiv);
            
            if (timestamp) {
                const timestampDiv = document.createElement('div');
                timestampDiv.className = 'timestamp';
                timestampDiv.textContent = timestamp;
                messageDiv.appendChild(timestampDiv);
            }
            
            chatBox.appendChild(messageDiv);
            chatBox.scrollTop = chatBox.scrollHeight;
        }

        function updateStatus(message) {
            document.getElementById('status').textContent = message;
        }

        function handleKeyPress(event) {
            if (event.key === 'Enter') {
                sendMessage();
            }
        }

        function sendMessage() {
            const input = document.getElementById('messageInput');
            const sendButton = document.getElementById('sendButton');
            const message = input.value.trim();
            
            if (!message || isConnected) {
                return;
            }

            // 添加用户消息
            addMessage(message, true, new Date().toLocaleString());
            input.value = '';
            sendButton.disabled = true;
            isConnected = true;
            
            updateStatus('正在连接...');

            // 创建SSE连接
            eventSource = new EventSource('/stream?message=' + encodeURIComponent(message));
            
            let aiMessageContent = '';
            let aiMessageDiv = null;

            eventSource.onopen = function() {
                updateStatus('已连接，等待响应...');
            };

            eventSource.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    
                    if (data.error) {
                        addMessage('错误: ' + data.error, false, data.timestamp);
                        closeConnection();
                        return;
                    }

                    if (data.content) {
                        if (!aiMessageDiv) {
                            // 创建AI消息容器
                            const chatBox = document.getElementById('chatBox');
                            aiMessageDiv = document.createElement('div');
                            aiMessageDiv.className = 'message ai-message';
                            
                            const contentDiv = document.createElement('div');
                            aiMessageDiv.appendChild(contentDiv);
                            
                            const timestampDiv = document.createElement('div');
                            timestampDiv.className = 'timestamp';
                            timestampDiv.textContent = data.timestamp;
                            aiMessageDiv.appendChild(timestampDiv);
                            
                            chatBox.appendChild(aiMessageDiv);
                        }
                        
                        // 更新内容
                        aiMessageContent += data.content;
                        aiMessageDiv.firstChild.textContent = aiMessageContent;
                        
                        // 滚动到底部
                        const chatBox = document.getElementById('chatBox');
                        chatBox.scrollTop = chatBox.scrollHeight;
                    }

                    if (data.finished) {
                        updateStatus('响应完成');
                        closeConnection();
                    }
                } catch (e) {
                    console.error('解析响应失败:', e);
                    addMessage('解析响应失败', false, new Date().toLocaleString());
                    closeConnection();
                }
            };

            eventSource.onerror = function(event) {
                console.error('SSE连接错误:', event);
                updateStatus('连接错误');
                closeConnection();
            };
        }

        function closeConnection() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
            }
            isConnected = false;
            document.getElementById('sendButton').disabled = false;
            setTimeout(() => updateStatus(''), 3000);
        }

        // 页面加载完成后的初始化
        window.onload = function() {
            updateStatus('准备就绪');
            setTimeout(() => updateStatus(''), 2000);
        };

        // 页面关闭时清理连接
        window.onbeforeunload = function() {
            closeConnection();
        };
    </script>
</body>
</html>
`
