package service

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strconv"
)

// BasicField 基本信息中的可配置字段：可增删、可开关、可选图标
type BasicField struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Value   string `json:"value"`
	Icon    string `json:"icon"`
	Visible bool   `json:"visible"`
}

// BasicInfo 基本信息：姓名/求职意向/头像固定，其余为字段列表
type BasicInfo struct {
	Name   string       `json:"name"`
	Intent string       `json:"intent"`
	Avatar string       `json:"avatar"`
	Fields []BasicField `json:"fields"`
}

// Theme 简历样式：配色 + 文字 + 布局
type Theme struct {
	Dark         string  `json:"dark"`         // 丝带标题底色
	Accent       string  `json:"accent"`       // 强调色
	FontCN       string  `json:"fontCN"`       // 中文字体 def|yahei|ping|source|song|hei|kai
	FontEN       string  `json:"fontEN"`       // 英文字体 def|system|arial|times|georgia|consolas
	FontSize     int     `json:"fontSize"`     // 正文字号 px 12-18
	LineHeight   float64 `json:"lineHeight"`   // 行间距 1.2-2.4
	HeaderLayout string  `json:"headerLayout"` // 头部布局 center|left|flat
	FieldStyle   string  `json:"fieldStyle"`   // 信息展示 icon|text|plain
}

// Item 模块内条目：通用键值，兼顾教育/实习/项目/荣誉/自定义
type Item struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Time     string   `json:"time"`
	Desc     string   `json:"desc"`
	Tags     []string `json:"tags"`
}

// Section 简历模块
type Section struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Visible bool   `json:"visible"`
	Content string `json:"content"`
	Items   []Item `json:"items"`
}

// RenderData 模板渲染上下文
type RenderData struct {
	Title     string
	BasicInfo BasicInfo
	Sections  []Section
	Theme     Theme
}

// 旧版扁平字段 → 新字段结构的迁移映射
var legacyFields = []struct{ key, label, icon string }{
	{"gender", "性别", "user"},
	{"age", "年龄", "calendar"},
	{"location", "所在城市", "map-pin"},
	{"phone", "手机号码", "phone"},
	{"email", "邮箱", "mail"},
}

// 常见字段的默认图标
var defaultIcons = map[string]string{
	"gender": "user", "age": "calendar", "location": "map-pin",
	"phone": "phone", "email": "mail", "political": "flag",
	"native": "map-pin", "birthday": "cake", "degree": "graduation-cap",
	"wechat": "message-circle", "github": "github", "homepage": "globe",
	"workyears": "briefcase",
}

func defaultIcon(key string) string {
	if icon, ok := defaultIcons[key]; ok {
		return icon
	}
	return "user"
}

// NormalizeBasic 将基本信息规整为字段列表结构，兼容旧版扁平字段，并补全缺省值
func NormalizeBasic(in json.RawMessage) json.RawMessage {
	if len(in) == 0 {
		in = json.RawMessage(`{}`)
	}
	var raw struct {
		Name   string          `json:"name"`
		Intent string          `json:"intent"`
		Avatar string          `json:"avatar"`
		Fields json.RawMessage `json:"fields"`
		Gender string          `json:"gender"`
		Age    string          `json:"age"`
		Loc    string          `json:"location"`
		Phone  string          `json:"phone"`
		Email  string          `json:"email"`
	}
	if err := json.Unmarshal(in, &raw); err != nil {
		return in
	}

	var fields []BasicField
	if len(raw.Fields) > 0 {
		_ = json.Unmarshal(raw.Fields, &fields)
	}
	if len(fields) == 0 {
		vals := map[string]string{
			"gender": raw.Gender, "age": raw.Age, "location": raw.Loc,
			"phone": raw.Phone, "email": raw.Email,
		}
		for _, lf := range legacyFields {
			fields = append(fields, BasicField{
				Key: lf.key, Label: lf.label, Icon: lf.icon,
				Value: vals[lf.key], Visible: true,
			})
		}
	}
	for i := range fields {
		f := &fields[i]
		if f.Key == "" {
			f.Key = "c_" + strconv.Itoa(i)
		}
		if f.Label == "" {
			f.Label = f.Key
		}
		if f.Icon == "" {
			f.Icon = defaultIcon(f.Key)
		}
	}
	out := map[string]any{
		"name": raw.Name, "intent": raw.Intent,
		"avatar": raw.Avatar, "fields": fields,
	}
	b, _ := json.Marshal(out)
	return b
}

func isHexColor(s string) bool {
	if len(s) != 7 && len(s) != 4 {
		return false
	}
	if s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// NormalizeTheme 校验并回填全部样式项，默认：蓝黑 + 系统字体 + 14px/1.6 + 居中 + 图标
func NormalizeTheme(in json.RawMessage) json.RawMessage {
	var t Theme
	if len(in) > 0 {
		_ = json.Unmarshal(in, &t)
	}
	if !isHexColor(t.Dark) {
		t.Dark = "#16213e"
	}
	if !isHexColor(t.Accent) {
		t.Accent = "#2563eb"
	}
	if !containsStr([]string{"def", "yahei", "ping", "source", "song", "hei", "kai"}, t.FontCN) {
		t.FontCN = "def"
	}
	if !containsStr([]string{"def", "system", "arial", "times", "georgia", "consolas"}, t.FontEN) {
		t.FontEN = "def"
	}
	if t.FontSize < 12 || t.FontSize > 18 {
		t.FontSize = 14
	}
	if t.LineHeight < 1.2 || t.LineHeight > 2.4 {
		t.LineHeight = 1.6
	}
	t.LineHeight = float64(int(t.LineHeight*100+0.5)) / 100
	if !containsStr([]string{"center", "left", "flat"}, t.HeaderLayout) {
		t.HeaderLayout = "center"
	}
	if !containsStr([]string{"icon", "text", "plain"}, t.FieldStyle) {
		t.FieldStyle = "icon"
	}
	b, _ := json.Marshal(t)
	return b
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// Parse 解析简历 JSON 为渲染数据
func Parse(basic, sections, theme json.RawMessage) (*RenderData, error) {
	d := &RenderData{}
	if len(basic) > 0 {
		if err := json.Unmarshal(NormalizeBasic(basic), &d.BasicInfo); err != nil {
			return nil, err
		}
	}
	if len(sections) > 0 {
		if err := json.Unmarshal(sections, &d.Sections); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal(NormalizeTheme(theme), &d.Theme); err != nil {
		return nil, err
	}
	visible := d.Sections[:0]
	for _, s := range d.Sections {
		if s.Visible {
			visible = append(visible, s)
		}
	}
	d.Sections = visible
	return d, nil
}

// splitLines 按换行拆分为多条（供 Markdown 渲染与模板共用）
func splitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

// iconSVGs 字段图标（lucide 线性风格，与简历纸张风格一致）
var iconSVGs = map[string]string{
	"user":           `<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle>`,
	"calendar":       `<path d="M8 2v4"></path><path d="M16 2v4"></path><rect width="18" height="18" x="3" y="4" rx="2"></rect><path d="M3 10h18"></path>`,
	"map-pin":        `<path d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"></path><circle cx="12" cy="10" r="3"></circle>`,
	"phone":          `<path d="M13.832 16.568a1 1 0 0 0 1.213-.303l.355-.465A2 2 0 0 1 17 15h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2A18 18 0 0 1 2 4a2 2 0 0 1 2-2h3a2 2 0 0 1 2 2v3a2 2 0 0 1-.8 1.6l-.468.351a1 1 0 0 0-.292 1.233 14 14 0 0 0 6.392 6.384"></path>`,
	"mail":           `<path d="m22 7-8.991 5.727a2 2 0 0 1-2.009 0L2 7"></path><rect x="2" y="4" width="20" height="16" rx="2"></rect>`,
	"message-circle": `<path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"></path>`,
	"globe":          `<circle cx="12" cy="12" r="10"></circle><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"></path><path d="M2 12h20"></path>`,
	"github":         `<path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4"></path><path d="M9 18c-4.51 2-5-2-7-2"></path>`,
	"flag":           `<path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"></path><line x1="4" x2="4" y1="22" y2="15"></line>`,
	"award":          `<path d="m15.477 12.89 1.515 8.526a.5.5 0 0 1-.81.47l-3.58-2.687a1 1 0 0 0-1.197 0l-3.586 2.686a.5.5 0 0 1-.81-.469l1.514-8.526"></path><circle cx="12" cy="8" r="6"></circle>`,
	"cake":           `<path d="M20 21v-8a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8"></path><path d="M4 16s.5-1 2-1 2.5 2 4 2 2.5-2 4-2 2.5 2 4 2 2-.5 2-1"></path><rect width="16" height="4" x="4" y="7" rx="1"></rect><path d="M12 7V4"></path><path d="M10 4h4"></path>`,
	"graduation-cap": `<path d="M21.42 10.922a1 1 0 0 0-.019-1.838L12.83 5.18a2 2 0 0 0-1.66 0L2.6 9.08a1 1 0 0 0 0 1.832l8.57 3.908a2 2 0 0 0 1.66 0z"></path><path d="M22 10v6"></path><path d="M6 12.5V16a6 3 0 0 0 12 0v-3.5"></path>`,
	"briefcase":      `<path d="M16 20V4a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path><rect width="20" height="14" x="2" y="6" rx="2"></rect>`,
	"link":           `<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>`,
}

// RenderHTML 用模板渲染简历页面（预览、PDF、Word 共用）
func RenderHTML(tmplSrc string, d *RenderData) (string, error) {
	t, err := template.New("resume").Funcs(template.FuncMap{
		"splitLines": splitLines,
		"richText":   func(s string) template.HTML { return template.HTML(RichHTML(s)) },
		"iconSVG": func(name string) template.HTML {
			paths, ok := iconSVGs[name]
			if !ok || name == "none" || name == "" {
				return ""
			}
			return template.HTML(`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#333333" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`)
		},
	}).Parse(tmplSrc)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}
