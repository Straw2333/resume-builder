package service

import "strings"

// RenderMarkdown 将简历渲染为 Markdown 文本，用于导出 .md
func RenderMarkdown(d *RenderData) string {
	var b strings.Builder
	bi := d.BasicInfo
	b.WriteString("# " + bi.Name + "\n\n")
	if bi.Intent != "" {
		b.WriteString("> 求职意向：" + bi.Intent + "\n\n")
	}
	var meta []string
	for _, f := range bi.Fields {
		if f.Visible && f.Value != "" {
			meta = append(meta, f.Label+"："+f.Value)
		}
	}
	if len(meta) > 0 {
		b.WriteString(strings.Join(meta, " ｜ ") + "\n\n")
	}

	for _, s := range d.Sections {
		b.WriteString("## " + s.Title + "\n\n")
		if s.Type == "evaluation" {
			for _, line := range splitLines(s.Content) {
				if line != "" {
					b.WriteString("- " + line + "\n")
				}
			}
			b.WriteString("\n")
			continue
		}
		for _, it := range s.Items {
			var head []string
			if it.Title != "" {
				head = append(head, "**"+it.Title+"**")
			}
			if it.Subtitle != "" {
				head = append(head, it.Subtitle)
			}
			if it.Time != "" {
				head = append(head, it.Time)
			}
			if len(head) > 0 {
				b.WriteString(strings.Join(head, " · ") + "\n\n")
			}
			wrote := false
			if it.Desc != "" {
				if md := RichTextToMarkdown(it.Desc); md != "" {
					b.WriteString(md + "\n")
					wrote = true
				}
			}
			if len(it.Tags) > 0 {
				b.WriteString("- 技能标签：" + strings.Join(it.Tags, "、") + "\n")
				wrote = true
			}
			if wrote {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}
