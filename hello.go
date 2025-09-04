package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NovitaTaskStatusResponse 定义Novita API响应结构
type NovitaTaskStatusResponse struct {
	Task struct {
		Status          string `json:"status"`
		Reason          string `json:"reason"`
		ProgressPercent int    `json:"progress_percent"`
	} `json:"task"`
	Images []struct {
		ImageUrl string `json:"image_url"`
	} `json:"images"`
}

// WebhookEvent 定义webhook事件结构
type WebhookEvent struct {
	EventType string `json:"event_type"`
	Payload   struct {
		Task struct {
			Eta             int    `json:"eta"`
			ProgressPercent int    `json:"progress_percent"`
			Reason          string `json:"reason"`
			Status          string `json:"status"`
			TaskId          string `json:"task_id"`
			TaskType        string `json:"task_type"`
		} `json:"task"`
		Extra struct {
			EnableNsfwDetection bool `json:"enable_nsfw_detection"`
			Webhook             struct {
				Url string `json:"url"`
			} `json:"webhook"`
		} `json:"extra"`
		Images []struct {
			ImageType           string      `json:"image_type"`
			ImageUrl            string      `json:"image_url"`
			ImageUrlTtl         string      `json:"image_url_ttl"`
			NsfwDetectionResult interface{} `json:"nsfw_detection_result"`
		} `json:"images"`
		Videos []interface{} `json:"videos"`
		Audios []interface{} `json:"audios"`
	} `json:"payload"`
}

// Txt2ImageRequest 定义文本转图片请求结构
type Txt2ImageRequest struct {
	ModelName string `json:"model_name"`
	Prompt    string `json:"prompt"`
	Extra     struct {
		Webhook struct {
			Url string `json:"url"`
		} `json:"webhook"`
	} `json:"extra"`
}

// Txt2ImageResponse 定义文本转图片响应结构
type Txt2ImageResponse struct {
	TaskId string `json:"task_id"`
}

// QueryTaskRequest 定义查询任务请求结构（带webhook）
type QueryTaskRequest struct {
	Extra struct {
		Webhook struct {
			Url string `json:"url"`
		} `json:"webhook"`
	} `json:"extra"`
}

func queryNovitaTaskStatus(ctx context.Context, taskId string, apiKey string, webhookUrl string) (int, string) {
	queryUrl := fmt.Sprintf("https://api.novita.ai/v3/async/task-result?task_id=%s", taskId)

	var httpReq *http.Request
	var err error

	// 创建包含webhook的请求体
	request := QueryTaskRequest{}
	request.Extra.Webhook.Url = webhookUrl

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("序列化请求数据失败: %v\n", err)
		return 1, ""
	}

	httpReq, err = http.NewRequestWithContext(ctx, "GET", queryUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Printf("创建HTTP请求失败: %v\n", err)
		return 1, ""
	}

	httpReq.Header.Set("Authorization", apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("发送HTTP请求失败: %v\n", err)
		return 1, ""
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		fmt.Printf("任务不存在，task_id: %s\n", taskId)
		return 3, ""
	}

	if httpResp.StatusCode != http.StatusOK {
		fmt.Printf("第三方API返回错误状态码: %d\n", httpResp.StatusCode)
		return 1, ""
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		fmt.Printf("读取响应内容失败: %v\n", err)
		return 1, ""
	}

	var novitaResp NovitaTaskStatusResponse
	if err := json.Unmarshal(respBody, &novitaResp); err != nil {
		fmt.Printf("解析Novita API响应失败: %v, 响应内容: %s\n", err, string(respBody))
		return 1, ""
	}

	var status int
	var imageUrl string

	switch novitaResp.Task.Status {
	case "TASK_STATUS_SUCCEED":
		status = 2 // 成功
		if len(novitaResp.Images) > 0 {
			imageUrl = novitaResp.Images[0].ImageUrl
		}
	case "TASK_STATUS_FAILED":
		status = 3 // 失败
		fmt.Printf("任务执行失败，task_id: %s, 原因: %s\n", taskId, novitaResp.Task.Reason)
	case "TASK_STATUS_PROCESSING", "TASK_STATUS_PENDING", "TASK_STATUS_RUNNING":
		status = 1 // 处理中
	default:
		fmt.Printf("未知的任务状态: %s, task_id: %s\n", novitaResp.Task.Status, taskId)
		status = 1 // 默认为处理中
	}

	statusText := map[int]string{1: "处理中", 2: "成功", 3: "失败"}
	fmt.Printf("查询任务状态，task_id: %s, 状态: %s, 进度: %d%%\n", taskId, statusText[status], novitaResp.Task.ProgressPercent)

	return status, imageUrl
}

// webhookHandler 处理webhook回调
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("读取webhook请求体失败: %v\n", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("收到webhook回调，原始数据: %s\n", string(body))

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		fmt.Printf("解析webhook数据失败: %v\n", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Printf("=== Webhook事件处理 ===\n")
	fmt.Printf("事件类型: %s\n", event.EventType)
	fmt.Printf("任务ID: %s\n", event.Payload.Task.TaskId)
	fmt.Printf("任务状态: %s\n", event.Payload.Task.Status)
	fmt.Printf("任务类型: %s\n", event.Payload.Task.TaskType)
	fmt.Printf("进度: %d%%\n", event.Payload.Task.ProgressPercent)

	if event.Payload.Task.Status == "TASK_STATUS_SUCCEED" {
		fmt.Printf("任务执行成功！\n")
		if len(event.Payload.Images) > 0 {
			fmt.Printf("生成的图片URL: %s\n", event.Payload.Images[0].ImageUrl)
			fmt.Printf("图片类型: %s\n", event.Payload.Images[0].ImageType)
			fmt.Printf("URL有效期: %s秒\n", event.Payload.Images[0].ImageUrlTtl)
		}
	} else if event.Payload.Task.Status == "TASK_STATUS_FAILED" {
		fmt.Printf("任务执行失败，原因: %s\n", event.Payload.Task.Reason)
	}

	// 返回200状态码表示webhook处理成功
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// startWebhookServer 启动webhook服务器
func startWebhookServer(port string) {
	http.HandleFunc("/webhook", webhookHandler)
	fmt.Printf("Webhook服务器启动在端口 %s\n", port)
	fmt.Printf("Webhook URL: http://localhost%s/webhook\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("启动webhook服务器失败: %v\n", err)
	}
}

func main() {
	go startWebhookServer(":8080")

	time.Sleep(2 * time.Second)

	ctx := context.Background()
	apiKey := ""
	webhookUrl := "http://localhost:8080/webhook"
	taskId := ""

	status, imageUrl := queryNovitaTaskStatus(ctx, taskId, apiKey, webhookUrl)
	fmt.Printf("任务状态: %d, 图片URL: %s\n", status, imageUrl)

	// 保持程序运行以接收webhook回调
	select {}
}
