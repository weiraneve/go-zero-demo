package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// 结构体定义
type Recollection struct {
	Entry   int    `json:"entry"`
	Content string `json:"content"`
	Pinned  bool   `json:"pinned"`
}

type Memories struct {
	Recollections []Recollection `json:"recollections"`
}

func main() {
	// 模拟 LLM 返回的原始内容（包含大量转义符）
	llmResponse := "{\"recollections\": [  \n  {\"entry\": 1, \"content\": \"A solitary figure stands on a weathered sidewalk in a quiet neighborhood, observing the surroundings with calculating eyes as the late afternoon sun casts long shadows.\", \"pinned\": false},  \n  {\"entry\": 2, \"content\": \"The figure murmurs, \\\"First impressions are everything,\\\" adjusting the sleeve of a well-worn jacket, as the world seems to hold its breath.\", \"pinned\": false},  \n  {\"entry\": 3, \"content\": \"A shrill whistle pierces the air, shattering the silence, and the figure's head snaps up, scanning the rooftops and alleyways for the source.\", \"pinned\": false},  \n  {\"entry\": 4, \"content\": \"A lean man in a dark suit jogs into view, saying, \\\"The boss is sending company. Time to let sleeping dogs lie,\\\" and nods to the first figure.\", \"pinned\": false},  \n  {\"entry\": 5, \"content\": \"The scent of approaching rain is carried by a gentle breeze that rustles through nearby trees.\", \"pinned\": false},  \n  {\"entry\": 6, \"content\": \"The figure and the man in the dark suit melt back into the shadows after the man says, \\\"The boss is sending company. Time to let sleeping dogs lie.\\\"\", \"pinned\": false},  \n  {\"entry\": 7, \"content\": \"The lone figure disappears into a dimly lit alley, the departing sound of footsteps quickly swallowed by the evening calm. A faint mist begins to rise from the sidewalk.\", \"pinned\": false},  \n  {\"entry\": 8, \"content\": \"As the last echoes of the departing footsteps fade, an eerie stillness settles over the deserted street. The distant hum of a lone car driving by and the soft rustle of leaves in the growing breeze are the only signs of life. Shadows stretch and yawn, growing longer and darker as night's veil slowly descends, wrapping the scene in an air of mystery and anticipation.\", \"pinned\": false}\n  ]\n}"

	fmt.Println("=== 原始 LLM 响应 ===")
	fmt.Printf("原始内容: %q\n\n", llmResponse)

	// 清理 LLM 响应
	cleanJSON := cleanLLMResponse(llmResponse)

	fmt.Println("=== 清理后的 JSON ===")
	fmt.Printf("清理后内容: %s\n\n", cleanJSON)

	// 验证是否为有效 JSON
	if isValidJSON(cleanJSON) {
		fmt.Println("✅ JSON 格式有效")

		// 解析 JSON
		var memories Memories
		err := json.Unmarshal([]byte(cleanJSON), &memories)
		if err != nil {
			fmt.Printf("❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("✅ 解析成功，共 %d 条记录\n\n", len(memories.Recollections))

			// 显示解析结果
			for i, rec := range memories.Recollections {
				fmt.Printf("记录 %d:\n", i+1)
				fmt.Printf("  Entry: %s\n", rec.Entry)
				fmt.Printf("  Content: %s\n", rec.Content)
				fmt.Printf("  Pinned: %t\n\n", rec.Pinned)
			}
		}
	} else {
		fmt.Println("❌ 清理后仍不是有效的 JSON")
	}
}

// 清理 LLM 响应的主函数
func cleanLLMResponse(response string) string {
	cleaned := response

	// 移除实际的换行符
	cleaned = strings.ReplaceAll(cleaned, "\n", "")

	// 移除多余的 ```json 和 ```
	if strings.HasPrefix(cleaned, "```json") && strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	return cleaned
}

// 修复常见的 JSON 格式问题
func fixCommonJSONIssues(jsonStr string) string {
	// 移除对象或数组结束前的多余逗号
	commaRegex := regexp.MustCompile(`,(\s*[}\]])`)
	jsonStr = commaRegex.ReplaceAllString(jsonStr, "$1")

	return jsonStr
}

// 验证是否为有效的 JSON
func isValidJSON(jsonStr string) bool {
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}
