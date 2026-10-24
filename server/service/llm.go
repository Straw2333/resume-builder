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

var llmHTTP = &http.Client{Timeout: 180 * time.Second}

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
	Instruction  string `json:"instruction"` // 用户指定的优化方向或追问修改要求
	CurrentHTML  string `json:"currentHtml"` // 多轮微调时上一轮的优化结果 HTML
}

const polishSystem = `你是一名资深的简历优化专家，负责润色求职者的简历内容。无需长篇推理思考，请直接快速输出润色结果。

要求：
1. 严格保留原文的事实，不得编造新的经历、项目、数字或成果；原文已有的数字必须原样保留。
2. 用简洁有力的语言突出成果与价值，动作开头（如：负责、主导、搭建、优化、提升），避免空话套话。
3. 使用中文输出。
4. 只输出润色后的内容本身，不要任何解释、前后缀，也不要 Markdown 代码块标记。
5. 输出为 HTML 片段，只允许这些标签：<ul> <ol> <li> <b> <strong> <i> <em> <u> <s> <br> <p>，禁止使用其他标签和任何属性。
6. 每个要点用 <li> 包裹组成 <ul>；重要关键词用 <b> 加粗，每条最多加粗 2-3 处。
7. 除非用户明确要求合并或增删要点，否则请保持输出 <li> 的条数、顺序与原文各要点一一对应，以便逐条比对。
8. 若用户提供了「优化方向 / 修改要求」，必须优先严格按用户的要求进行针对性调整。`

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
	if cur := strings.TrimSpace(in.CurrentHTML); cur != "" {
		sb.WriteString("\n\n上一版 AI 优化结果：\n" + cur)
	}
	if inst := strings.TrimSpace(in.Instruction); inst != "" {
		if strings.TrimSpace(in.CurrentHTML) != "" {
			sb.WriteString("\n\n用户的追问修改要求（请在上一版优化结果的基础上，结合原文事实进行针对性修改）：\n" + inst)
		} else {
			sb.WriteString("\n\n用户的优化方向要求（请重点遵循）：\n" + inst)
		}
	}

	out, err := Chat(cfg, polishSystem, sb.String())
	if err != nil {
		return "", err
	}
	// 剥离推理模型可能附带的 <think>...</think> 思考块与 ```html 围栏，再净化
	if idx := strings.Index(out, "</think>"); idx >= 0 {
		out = out[idx+len("</think>"):]
	}
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

// ===== 根据 JD 自动生成简历 =====

// GenerateFromJDInput 根据招聘 JD 生成简历的输入参数。
type GenerateFromJDInput struct {
	JD          string `json:"jd"`          // 招聘 JD 原文（必填）
	UserBg      string `json:"userBg"`      // 个人真实背景/侧重点补充（选填）
	Instruction string `json:"instruction"` // 对话微调指令（选填）
	CurrentJSON string `json:"currentJson"` // 多轮微调时上一版生成的简历 JSON（选填）
}

// GeneratedResume AI 根据 JD 生成的结构化简历数据。
type GeneratedResume struct {
	Title    string    `json:"title"`
	Intent   string    `json:"intent"`
	Sections []Section `json:"sections"`
}

const jdGenerateSystem = `你是一名资深的技术与职场简历架构师。请根据用户提供的「招聘岗位 JD」以及可选的「个人背景补充」，量身生成一份高匹配度、专业且结构完整的中文简历草稿。无需长篇推理思考，请直接快速输出纯 JSON 结果。

要求：
1. 深度提取 JD 中的岗位名称、核心技术栈、业务场景与职责要求，使生成的专业技能、项目经历、实习经历与 JD 高度契合。
2. 若用户提供了个人背景信息（如院校、专业、掌握技术、做过的项目或实习），严格以用户真实信息为基础展开；对于用户未提供的具体名称或量化指标，提供贴合岗位的专业范本（可在括号或加粗处温和提示用户按实际微调，如“某某科技有限公司（请填实际公司）”）。
3. 只输出一个合法的 JSON 对象，不要任何解释文字，不要 Markdown 代码块标记。
4. JSON 结构必须严格符合以下格式：
{
  "title": "简历标题（如：Go后端开发工程师-目标简历）",
  "intent": "求职意向岗位名（提取自JD）",
  "sections": [
    {
      "type": "education",
      "title": "教育背景",
      "visible": true,
      "content": "",
      "items": [
        {
          "title": "学校名称",
          "subtitle": "专业 / 学历",
          "time": "2021.09 - 2025.06",
          "desc": "<ul><li>主修课程与学业表现要点，重要词汇用 <b> 加粗</li></ul>",
          "tags": []
        }
      ]
    },
    {
      "type": "skill",
      "title": "专业技能",
      "visible": true,
      "content": "",
      "items": [
        {
          "title": "技能分类名称",
          "subtitle": "熟练度（如：熟练掌握）",
          "time": "",
          "desc": "<ul><li>具体技能要点 1，关键词用 <b> 加粗</li><li>具体技能要点 2</li></ul>",
          "tags": ["技术标签1", "技术标签2"]
        }
      ]
    },
    {
      "type": "internship",
      "title": "实习经历",
      "visible": true,
      "content": "",
      "items": [
        {
          "title": "公司名称",
          "subtitle": "实习岗位",
          "time": "2024.06 - 2024.09",
          "desc": "<ul><li>按 STAR 法则描述工作职责与量化成果，每条一个 <li>，核心技术与指标用 <b> 加粗</li></ul>",
          "tags": ["相关技术栈"]
        }
      ]
    },
    {
      "type": "project",
      "title": "项目经历",
      "visible": true,
      "content": "",
      "items": [
        {
          "title": "贴合 JD 场景的项目名称",
          "subtitle": "担任角色（如：核心开发）",
          "time": "2024.03 - 2024.06",
          "desc": "<ul><li>项目背景与架构设计要点</li><li>核心技术攻关与性能优化量化成果</li></ul>",
          "tags": ["技术栈1", "技术栈2"]
        }
      ]
    },
    {
      "type": "evaluation",
      "title": "自我评价",
      "visible": true,
      "content": "第一条贴合岗位要求的自我评价\n第二条自我评价\n第三条自我评价",
      "items": []
    }
  ]
}
5. 注意：列表型模块（education/skill/internship/project/honor）的 desc 字段必须是 <ul><li>...</li></ul> 格式的 HTML 字符串，只允许 ul/ol/li/b/strong/i 标签；evaluation 模块的 content 字段为按换行符 \n 分隔的多行纯文本。
6. 若用户提供了「上一版生成草稿」和「追问修改要求」，请在上一版草稿基础上严格按修改要求进行针对性调整，并返回完整 JSON。`

// GenerateFromJD 根据 JD（及可选个人背景、追问指令）生成结构化简历。
func GenerateFromJD(cfg *AISettings, in *GenerateFromJDInput) (*GeneratedResume, error) {
	if strings.TrimSpace(in.JD) == "" {
		return nil, fmt.Errorf("请先粘贴招聘岗位 JD 内容")
	}
	var sb strings.Builder
	sb.WriteString("请根据以下招聘岗位 JD 生成匹配的简历内容：\n\n【岗位 JD】\n" + strings.TrimSpace(in.JD))
	if bg := strings.TrimSpace(in.UserBg); bg != "" {
		sb.WriteString("\n\n【求职者个人背景 / 定制要求】\n" + bg)
	}
	if cur := strings.TrimSpace(in.CurrentJSON); cur != "" {
		sb.WriteString("\n\n【上一版生成草稿 JSON】\n" + cur)
	}
	if inst := strings.TrimSpace(in.Instruction); inst != "" {
		sb.WriteString("\n\n【本轮修改指令（请严格遵循）】\n" + inst)
	}

	out, err := Chat(cfg, jdGenerateSystem, sb.String())
	if err != nil {
		return nil, err
	}
	if idx := strings.Index(out, "</think>"); idx >= 0 {
		out = out[idx+len("</think>"):]
	}
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```json")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)
	if start, end := strings.IndexByte(out, '{'), strings.LastIndexByte(out, '}'); start >= 0 && end > start {
		out = out[start : end+1]
	}

	var res GeneratedResume
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		// 容错：部分模型会在 JSON 字符串值内部直接输出未转义换行符，修复后再试一次
		fixed := escapeMultilineJSONStrings(out)
		if err2 := json.Unmarshal([]byte(fixed), &res); err2 != nil {
			return nil, fmt.Errorf("AI 生成结果解析失败，请重试: %w", err)
		}
	}
	normalizeGeneratedResume(&res)
	if len(res.Sections) == 0 {
		return nil, fmt.Errorf("AI 未生成有效的简历模块，请重试")
	}
	return &res, nil
}

// escapeMultilineJSONStrings 将 JSON 双引号字符串内部未转义的真实换行/制表符转为 \n \t
func escapeMultilineJSONStrings(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	inStr := false
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			sb.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' && inStr {
			sb.WriteByte(c)
			escaped = true
			continue
		}
		if c == '"' {
			inStr = !inStr
			sb.WriteByte(c)
			continue
		}
		if inStr {
			switch c {
			case '\n':
				sb.WriteString(`\n`)
				continue
			case '\r':
				continue
			case '\t':
				sb.WriteString(`\t`)
				continue
			}
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

// normalizeGeneratedResume 净化并补齐生成的简历模块结构，确保与前端编辑器 100% 兼容。
func normalizeGeneratedResume(res *GeneratedResume) {
	if strings.TrimSpace(res.Intent) == "" {
		res.Intent = "目标岗位"
	}
	if strings.TrimSpace(res.Title) == "" {
		res.Title = res.Intent + "-简历"
	}

	defaultMeta := []struct {
		typ   string
		title string
	}{
		{"education", "教育背景"},
		{"skill", "专业技能"},
		{"internship", "实习经历"},
		{"project", "项目经历"},
		{"honor", "荣誉证书"},
		{"evaluation", "自我评价"},
		{"custom", "自定义模块"},
	}

	byType := make(map[string]Section)
	var extraSections []Section

	for _, s := range res.Sections {
		s.Type = strings.TrimSpace(s.Type)
		if s.Type == "evaluation" {
			s.Title = nonEmpty(s.Title, "自我评价")
			// 若模型误将自我评价放入 items，提取合并回 content
			if strings.TrimSpace(s.Content) == "" && len(s.Items) > 0 {
				var lines []string
				for _, it := range s.Items {
					if t := strings.TrimSpace(RichTextToMarkdown(it.Desc)); t != "" {
						lines = append(lines, t)
					} else if t := strings.TrimSpace(it.Title); t != "" {
						lines = append(lines, t)
					}
				}
				s.Content = strings.Join(lines, "\n")
			}
			// 剥去可能混入的列表符号或 HTML
			if strings.Contains(s.Content, "<") {
				var cleanLines []string
				for _, ln := range splitLines(RichTextToMarkdown(s.Content)) {
					if c := stripListMarker(ln); c != "" {
						cleanLines = append(cleanLines, c)
					}
				}
				s.Content = strings.Join(cleanLines, "\n")
			}
			s.Items = []Item{}
			s.Visible = strings.TrimSpace(s.Content) != ""
			byType["evaluation"] = s
			continue
		}

		var cleanItems []Item
		for _, it := range s.Items {
			it.Title = strings.TrimSpace(it.Title)
			it.Subtitle = strings.TrimSpace(it.Subtitle)
			it.Time = strings.TrimSpace(it.Time)
			it.Desc = RichHTML(it.Desc)
			if it.Tags == nil {
				it.Tags = []string{}
			}
			if it.Title != "" || it.Desc != "" {
				cleanItems = append(cleanItems, it)
			}
		}
		s.Items = cleanItems
		s.Visible = len(cleanItems) > 0
		if _, exists := byType[s.Type]; !exists && isStandardSectionType(s.Type) {
			byType[s.Type] = s
		} else if len(cleanItems) > 0 {
			extraSections = append(extraSections, s)
		}
	}

	var finalSections []Section
	for _, dm := range defaultMeta {
		if sec, ok := byType[dm.typ]; ok {
			sec.Title = nonEmpty(sec.Title, dm.title)
			finalSections = append(finalSections, sec)
		} else {
			finalSections = append(finalSections, Section{
				Type:    dm.typ,
				Title:   dm.title,
				Visible: false,
				Content: "",
				Items:   []Item{},
			})
		}
	}
	finalSections = append(finalSections, extraSections...)
	res.Sections = finalSections
}

func isStandardSectionType(t string) bool {
	switch t {
	case "education", "skill", "internship", "project", "honor", "evaluation", "custom":
		return true
	}
	return false
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}
