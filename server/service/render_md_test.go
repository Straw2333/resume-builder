package service

import (
	"strings"
	"testing"
)

func TestMarkdownTemplateRoundTrip(t *testing.T) {
	tmpl := MarkdownTemplate()
	title, basicJSON, sectionsJSON, err := ParseMarkdown(tmpl)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}
	if title != "Go 后端开发工程师-张三" {
		t.Errorf("unexpected title: %s", title)
	}
	d, err := Parse(basicJSON, sectionsJSON, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if d.BasicInfo.Name != "张三" || d.BasicInfo.Intent != "Go 后端开发工程师" {
		t.Errorf("basicInfo mismatch: %+v", d.BasicInfo)
	}
	if len(d.Sections) != 6 {
		t.Fatalf("expected 6 visible sections, got %d", len(d.Sections))
	}

	// 导出为 Markdown 再重新导入，验证双向闭环
	exported := RenderMarkdown(d)
	if !strings.Contains(exported, "### 某某科技有限公司 · 后端开发实习生 · 2023.06 - 2023.12") {
		t.Errorf("exported markdown missing item header:\n%s", exported)
	}
	if !strings.Contains(exported, "**核心链路优化**") {
		t.Errorf("exported markdown missing bold text:\n%s", exported)
	}

	title2, basic2, sections2, err := ParseMarkdown(exported)
	if err != nil {
		t.Fatalf("second ParseMarkdown failed: %v", err)
	}
	d2, err := Parse(basic2, sections2, nil)
	if err != nil {
		t.Fatalf("second Parse failed: %v", err)
	}
	if title2 != title || len(d2.Sections) != len(d.Sections) {
		t.Errorf("round-trip mismatch: title=%s/%s, sections=%d/%d", title, title2, len(d.Sections), len(d2.Sections))
	}
}

func TestLegacyExportedMarkdownImport(t *testing.T) {
	// 模拟旧版导出的 Markdown（条目头部使用 **Title** · Subtitle · Time）
	legacyMD := `# 李四

> 求职意向：前端开发工程师

性别：女 ｜ 年龄：23 ｜ 手机号码：13900001111

## 实习经历

**字节跳动** · 前端实习生 · 2024.01 - 2024.06

- 负责核心组件库维护，提升构建速度 **40%**
- 技能标签：Vue3、TypeScript、Vite
`
	title, basicJSON, sectionsJSON, err := ParseMarkdown(legacyMD)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}
	if title != "前端开发工程师-李四" {
		t.Errorf("unexpected title: %s", title)
	}
	d, err := Parse(basicJSON, sectionsJSON, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(d.Sections) != 1 || d.Sections[0].Type != "internship" {
		t.Fatalf("unexpected sections: %+v", d.Sections)
	}
	it := d.Sections[0].Items[0]
	if it.Title != "字节跳动" || it.Subtitle != "前端实习生" || it.Time != "2024.01 - 2024.06" {
		t.Errorf("item header mismatch: %+v", it)
	}
	if len(it.Tags) != 3 || it.Tags[0] != "Vue3" {
		t.Errorf("item tags mismatch: %+v", it.Tags)
	}
	if !strings.Contains(it.Desc, "<b>40%</b>") {
		t.Errorf("expected bold HTML in desc, got: %s", it.Desc)
	}
}

func TestEdgeCasesAndCustomSections(t *testing.T) {
	md := `# 王五
> 求职意向：架构师 | 期望薪资：40k
- **手机：**13700002222
- 个人主页：https://example.com

## 教育背景
### 清华大学 | 2016.09 - 2020.06
- 计算机系

## 开源贡献
### Gin 框架 Contributor | 核心维护者 | 2022 - 至今
1. 主导路由性能优化
  - 减少内存分配 **50%**
2. 修复关键并发问题
`
	_, basicJSON, sectionsJSON, err := ParseMarkdown(md)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}
	d, err := Parse(basicJSON, sectionsJSON, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if d.BasicInfo.Intent != "架构师" {
		t.Errorf("expected intent 架构师, got %s", d.BasicInfo.Intent)
	}
	// 检查两段式教育背景（学校 + 时间，无专业）
	edu := d.Sections[0].Items[0]
	if edu.Title != "清华大学" || edu.Subtitle != "" || edu.Time != "2016.09 - 2020.06" {
		t.Errorf("2-part time header parse failed: %+v", edu)
	}
	// 检查自定义模块与嵌套列表
	custom := d.Sections[1]
	if custom.Type != "custom" || custom.Title != "开源贡献" {
		t.Errorf("custom section parse failed: %+v", custom)
	}
	if !strings.Contains(custom.Items[0].Desc, "<ol>") || !strings.Contains(custom.Items[0].Desc, "<ul>") {
		t.Errorf("nested list HTML failed: %s", custom.Items[0].Desc)
	}
}
