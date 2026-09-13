package service

import (
	"html"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
	atom "golang.org/x/net/html/atom"
)

// 富文本白名单：仅保留简历描述所需的行内/列表标签，属性全部剥离
var richAllowed = map[string]bool{
	"b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"br": true, "p": true, "ul": true, "ol": true, "li": true,
}
var richDrop = map[string]bool{
	"script": true, "style": true, "meta": true, "link": true, "title": true,
}

// RichHTML 将描述规整为安全的富文本 HTML。
// 纯文本（旧数据，\n 分行）转为无序列表；HTML 输入只保留白名单标签。
func RichHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "<") {
		var sb strings.Builder
		sb.WriteString("<ul>")
		for _, l := range splitLines(s) {
			if t := strings.TrimSpace(l); t != "" {
				sb.WriteString("<li>" + html.EscapeString(t) + "</li>")
			}
		}
		sb.WriteString("</ul>")
		s = sb.String()
	}
	doc, err := xhtml.ParseFragment(strings.NewReader(s), &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return ""
	}
	var sb strings.Builder
	for _, n := range doc {
		sanitizeRichNode(n, &sb)
	}
	return sb.String()
}

func sanitizeRichNode(n *xhtml.Node, sb *strings.Builder) {
	switch n.Type {
	case xhtml.TextNode:
		sb.WriteString(html.EscapeString(n.Data))
	case xhtml.ElementNode:
		if richDrop[n.Data] {
			return
		}
		if richAllowed[n.Data] {
			if n.Data == "br" {
				sb.WriteString("<br/>")
				return
			}
			sb.WriteString("<" + n.Data + ">")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				sanitizeRichNode(c, sb)
			}
			sb.WriteString("</" + n.Data + ">")
		} else {
			// 非白名单标签剥壳保留内容
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				sanitizeRichNode(c, sb)
			}
		}
	}
}

// RichTextToMarkdown 富文本转 Markdown（加粗 **、有序/无序列表、嵌套缩进）；
// 纯文本按行转 "- " 列表，保持旧数据行为。
func RichTextToMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "<") {
		var out []string
		for _, l := range splitLines(s) {
			if t := strings.TrimSpace(l); t != "" {
				out = append(out, "- "+t)
			}
		}
		return strings.Join(out, "\n")
	}
	doc, err := xhtml.ParseFragment(strings.NewReader(s), &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return ""
	}
	var sb strings.Builder
	mdBlocks(doc, 0, &sb)
	return strings.TrimRight(sb.String(), "\n")
}

func mdBlocks(nodes []*xhtml.Node, indent int, sb *strings.Builder) {
	pad := strings.Repeat("  ", indent)
	for _, n := range nodes {
		if n.Type == xhtml.TextNode {
			if t := strings.TrimSpace(n.Data); t != "" {
				sb.WriteString(pad + t + "\n")
			}
			continue
		}
		if n.Type != xhtml.ElementNode {
			continue
		}
		switch n.Data {
		case "ul", "ol":
			counter := 1
			for li := n.FirstChild; li != nil; li = li.NextSibling {
				if li.Type != xhtml.ElementNode || li.Data != "li" {
					continue
				}
				prefix := pad + "- "
				if n.Data == "ol" {
					prefix = pad + strconv.Itoa(counter) + ". "
					counter++
				}
				sb.WriteString(prefix + mdInlineExceptLists(li) + "\n")
				for sub := li.FirstChild; sub != nil; sub = sub.NextSibling {
					if sub.Type == xhtml.ElementNode && (sub.Data == "ul" || sub.Data == "ol") {
						mdBlocks([]*xhtml.Node{sub}, indent+1, sb)
					}
				}
			}
		case "br":
			// 顶层孤立 br 忽略
		default:
			if richDrop[n.Data] {
				continue
			}
			if t := strings.TrimSpace(mdInlineExceptLists(n)); t != "" {
				sb.WriteString(pad + t + "\n")
			}
		}
	}
}

func mdInlineExceptLists(n *xhtml.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == xhtml.ElementNode && (c.Data == "ul" || c.Data == "ol") {
			continue
		}
		sb.WriteString(mdInline(c))
	}
	return sb.String()
}

func mdInline(n *xhtml.Node) string {
	switch n.Type {
	case xhtml.TextNode:
		return n.Data
	case xhtml.ElementNode:
		switch n.Data {
		case "br":
			return "\n"
		case "b", "strong":
			if inner := strings.TrimSpace(mdInlineAll(n)); inner != "" {
				return "**" + inner + "**"
			}
			return ""
		case "i", "em":
			if inner := strings.TrimSpace(mdInlineAll(n)); inner != "" {
				return "*" + inner + "*"
			}
			return ""
		case "ul", "ol", "script", "style":
			return ""
		}
		return mdInlineAll(n)
	}
	return ""
}

func mdInlineAll(n *xhtml.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(mdInline(c))
	}
	return sb.String()
}
