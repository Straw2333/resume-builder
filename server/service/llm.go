package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AISettings AI 服务配置，存于 settings 表（ai_settings），前端界面可配置。
// Provider 决定请求协议：openai = OpenAI 兼容 /chat/completions（含 DeepSeek、智谱等），
// anthropic = Claude /v1/messages。
type AISettings struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"baseURL"`
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
}

func (s *AISettings) Validate() error {
	if s.APIKey == "" {
		return fmt.Errorf("未配置 API Key")
	}
	if s.BaseURL == "" || s.Model == "" {
		return fmt.Errorf("AI 服务配置不完整，请在 AI 设置中补全")
	}
	if s.Provider != "openai" && s.Provider != "anthropic" {
		return fmt.Errorf("不支持的接口协议: %s", s.Provider)
	}
	return nil
}

var llmHTTP = &http.Client{Timeout: 110 * time.Second}

// Chat 调用 LLM 补全，system 为系统提示词，user 为用户消息，返回模型输出文本。
func Chat(cfg *AISettings, system, user string) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	switch cfg.Provider {
	case "anthropic":
		return chatAnthropic(cfg, system, user)
	default:
		return chatOpenAI(cfg, system, user)
	}
}

func chatOpenAI(cfg *AISettings, system, user string) (string, error) {
	base := strings.TrimRight(cfg.BaseURL, "/")
	url := base + "/chat/completions"
	body, _ := json.Marshal(map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	return doLLM(req, "openai")
}

func chatAnthropic(cfg *AISettings, system, user string) (string, error) {
	base := strings.TrimRight(cfg.BaseURL, "/")
	url := base + "/v1/messages"
	if strings.HasSuffix(base, "/v1") {
		url = base + "/messages"
	}
	body, _ := json.Marshal(map[string]any{
		"model":      cfg.Model,
		"max_tokens": 4096,
		"system":     system,
		"messages":   []map[string]string{{"role": "user", "content": user}},
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return doLLM(req, "anthropic")
}

// doLLM 发请求、解析两种协议的响应。status != 200 时尽力提取错误信息。
func doLLM(req *http.Request, protocol string) (string, error) {
	resp, err := llmHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("无法连接 AI 服务: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		msg := extractAPIError(protocol, raw)
		return "", fmt.Errorf("AI 服务返回错误 (HTTP %d)%s", resp.StatusCode, msg)
	}
	switch protocol {
	case "anthropic":
		var r struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return "", fmt.Errorf("AI 响应解析失败: %w", err)
		}
		var sb strings.Builder
		for _, c := range r.Content {
			sb.WriteString(c.Text)
		}
		if sb.Len() == 0 {
			return "", fmt.Errorf("AI 未返回内容")
		}
		return sb.String(), nil
	default:
		var r struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return "", fmt.Errorf("AI 响应解析失败: %w", err)
		}
		if len(r.Choices) == 0 {
			return "", fmt.Errorf("AI 未返回内容")
		}
		return r.Choices[0].Message.Content, nil
	}
}

func extractAPIError(protocol string, raw []byte) string {
	var e struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &e) != nil || len(e.Error) == 0 {
		return ""
	}
	// OpenAI 的 error 是对象 {"message": ...}，Anthropic 也是对象，但个别网关直接回字符串
	var obj struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(e.Error, &obj) == nil && obj.Message != "" {
		return ": " + obj.Message
	}
	var s string
	if json.Unmarshal(e.Error, &s) == nil && s != "" {
		return ": " + s
	}
	return ""
}

// ===== 简历润色 =====

// PolishInput 单条润色的输入。Content 为富文本 HTML（条目描述）或纯文本（自我评价，按行）。
type PolishInput struct {
	Kind         string `json:"kind"` // item-desc | section-content
	SectionType  string `json:"sectionType"`
	SectionTitle string `json:"sectionTitle"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	TargetRole   string `json:"targetRole"`
	Content      string `json:"content"`
}

const polishSystem = `你是一名资深的简历优化专家，负责润色求职者的简历内容。请将用户提供的简历片段优化得更专业、更有竞争力。

要求：
1. 严格保留原文的事实，不得编造新的经历、项目、数字或成果；原文已有的数字必须原样保留。
2. 用简洁有力的语言突出成果与价值，动作开头（如：负责、主导、搭建、优化、提升），避免空话套话。
3. 使用中文输出。
4. 只输出润色后的内容本身，不要任何解释、前后缀，也不要 Markdown 代码块标记。
5. 输出为 HTML 片段，只允许这些标签：<ul> <ol> <li> <b> <strong> <i> <em> <u> <s> <br> <p>，禁止使用其他标签和任何属性。
6. 每个要点用 <li> 包裹组成 <ul>；重要关键词用 <b> 加粗，每条最多加粗 2-3 处。`

// Polish 调用 LLM 润色简历片段，返回净化后的富文本 HTML（可直接写入 desc/content 存储与渲染）。
func Polish(cfg *AISettings, in *PolishInput) (string, error) {
	if strings.TrimSpace(in.Content) == "" {
		return "", fmt.Errorf("内容为空，无需优化")
	}
	var sb strings.Builder
	sb.WriteString("请优化以下「" + in.SectionTitle + "」模块中的内容。")
	if t := strings.TrimSpace(in.Title); t != "" {
		sb.WriteString("\n条目：" + t)
	}
	if t := strings.TrimSpace(in.Subtitle); t != "" {
		sb.WriteString("（" + t + "）")
	}
	if r := strings.TrimSpace(in.TargetRole); r != "" {
		sb.WriteString("\n目标岗位：" + r + "，请让内容更贴合该岗位的要求。")
	}
	if in.Kind == "section-content" {
		sb.WriteString("\n这是自我评价，原文每行一条，请保持逐条输出的结构。")
	}
	sb.WriteString("\n\n原文：\n" + in.Content)

	out, err := Chat(cfg, polishSystem, sb.String())
	if err != nil {
		return "", err
	}
	// 模型偶尔无视约束输出 ```html 围栏，剥掉后再净化
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```html")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)
	html := RichHTML(out)
	if html == "" {
		return "", fmt.Errorf("AI 返回了无法解析的内容，请重试")
	}
	return html, nil
}
