package service

import (
	"html"
	"strconv"
	"strings"
	"unicode"

	xhtml "golang.org/x/net/html"
	atom "golang.org/x/net/html/atom"
)

// 富文本白名单：仅保留简历描述所需的行内/列表标签，属性全部剥离
var richAllowed = map[string]bool{
	"b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"br": true, "p": true, "ul": true, "ol": true, "li": true,
}
var richDrop = map[string]bool{
	"script": true, "style": true, "meta": true, "link": true, "title": true, "think": true,
}

// RichHTML 将描述规整为安全的富文本 HTML。
// 纯文本输入（旧数据或手工输入的 Markdown 语法）按 Markdown 解析为列表/加粗；
// HTML 输入只保留白名单标签。
func RichHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "<") {
		return mdToHTML(s)
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

// ===== 简版 Markdown 解析 =====
// 支持与编辑器输入一致的语法：- * + 无序列表、1. 1、 1) 有序列表、
// 两个空格（或一个 Tab）缩进嵌套、**加粗**、*斜体*；无标记的行并入无序列表。

// orderedMarkerLen 检测行首有序列表标记（如 1. 1、 1)）的字节长度，无则返回 0
func orderedMarkerLen(t string) int {
	r := []rune(t)
	i := 0
	for i < len(r) && i < 3 && r[i] >= '0' && r[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(r) {
		return 0
	}
	if r[i] == '.' || r[i] == ')' || r[i] == '、' {
		i++
		if i < len(r) && unicode.IsSpace(r[i]) {
			i++
		}
		return len(string(r[:i]))
	}
	return 0
}

// firstRuneIsSpace 判断字符串首字符是否为空白
func firstRuneIsSpace(s string) bool {
	for _, r := range s {
		return unicode.IsSpace(r)
	}
	return false
}

// stripListMarker 去除行首 Markdown 列表标记，返回剩余内容（供按行渲染处共用）
func stripListMarker(s string) string {
	t := strings.TrimLeft(s, " \t")
	if len(t) >= 2 && (t[0] == '-' || t[0] == '*' || t[0] == '+') && firstRuneIsSpace(t[1:]) {
		return strings.TrimSpace(t[1:])
	}
	if k := orderedMarkerLen(t); k > 0 {
		return strings.TrimSpace(t[k:])
	}
	return strings.TrimSpace(t)
}

// mdInlineText 行内 **加粗** / *斜体* 转换，其余内容 HTML 转义
func mdInlineText(s string) string {
	var sb strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) && (s[i+1] == '*' || s[i+1] == '\\') {
			sb.WriteString(html.EscapeString(s[i+1 : i+2]))
			i += 2
			continue
		}
		if s[i] == '*' && i+2 < len(s) && s[i+1] == '*' {
			if j := strings.Index(s[i+2:], "**"); j >= 0 {
				sb.WriteString("<b>" + mdInlineText(s[i+2:i+2+j]) + "</b>")
				i += j + 4
				continue
			}
		}
		if s[i] == '*' {
			if j := strings.IndexByte(s[i+1:], '*'); j > 0 {
				if inner := s[i+1 : i+1+j]; !strings.Contains(inner, "*") {
					sb.WriteString("<i>" + mdInlineText(inner) + "</i>")
					i += j + 2
					continue
				}
			}
		}
		j := strings.IndexAny(s[i+1:], "*\\")
		if j < 0 {
			j = len(s) - i
		} else {
			j += 1
		}
		sb.WriteString(html.EscapeString(s[i : i+j]))
		i += j
	}
	return sb.String()
}

// mdToHTML 把纯文本描述按 Markdown 规则解析为白名单内的富文本 HTML；
// 全部文本经转义，输出标签仅限 ul/ol/li/b/i，无需再净化。
func mdToHTML(s string) string {
	type mline struct {
		level   int
		ordered bool
		html    string
	}
	var rows []mline
	for _, raw := range splitLines(s) {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		body := strings.TrimLeft(raw, " \t")
		lead := raw[:len(raw)-len(body)]
		level := len(strings.ReplaceAll(lead, "\t", "  ")) / 2
		if level > 6 {
			level = 6
		}
		ordered := false
		if len(body) >= 2 && (body[0] == '-' || body[0] == '*' || body[0] == '+') && firstRuneIsSpace(body[1:]) {
			body = strings.TrimLeft(body[1:], " \t")
		} else if k := orderedMarkerLen(body); k > 0 {
			ordered = true
			body = body[k:]
		}
		rows = append(rows, mline{level, ordered, mdInlineText(strings.TrimSpace(body))})
	}
	if len(rows) == 0 {
		return ""
	}
	type frame struct {
		level   int
		ordered bool
		open    bool
	}
	tag := func(ordered bool) string {
		if ordered {
			return "ol"
		}
		return "ul"
	}
	var sb strings.Builder
	var stack []frame
	closeTop := func() {
		f := &stack[len(stack)-1]
		if f.open {
			sb.WriteString("</li>")
			f.open = false
		}
		sb.WriteString("</" + tag(f.ordered) + ">")
		stack = stack[:len(stack)-1]
	}
	for _, ln := range rows {
		for len(stack) > 0 && stack[len(stack)-1].level > ln.level {
			closeTop()
		}
		if len(stack) > 0 && stack[len(stack)-1].level == ln.level && stack[len(stack)-1].ordered != ln.ordered {
			closeTop()
		}
		if len(stack) == 0 || stack[len(stack)-1].level < ln.level {
			sb.WriteString("<" + tag(ln.ordered) + ">")
			stack = append(stack, frame{level: ln.level, ordered: ln.ordered})
		}
		top := &stack[len(stack)-1]
		if top.open {
			sb.WriteString("</li>")
		}
		sb.WriteString("<li>" + ln.html)
		top.open = true
	}
	for len(stack) > 0 {
		closeTop()
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
// 纯文本输入先按 Markdown 解析再输出，保持标记与内容的对应。
func RichTextToMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "<") {
		return strings.TrimRight(mdToMarkdownStruct(s), "\n")
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
			for _, ln := range splitLines(n.Data) {
				if t := stripListMarker(ln); t != "" {
					sb.WriteString(pad + "- " + t + "\n")
				}
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
				if li.Type != xhtml.ElementNode {
					continue
				}
				if li.Data == "ul" || li.Data == "ol" {
					// 与 li 平级的嵌套列表（编辑器常见结构），按缩进继续展开，避免丢失
					mdBlocks([]*xhtml.Node{li}, indent+1, sb)
					continue
				}
				if li.Data != "li" {
					continue
				}
				prefix := pad + "- "
				if n.Data == "ol" {
					prefix = pad + strconv.Itoa(counter) + ". "
					counter++
				}
				// 空列表项（<li><br></li> 之类）跳过输出，避免导出孤立 "- " 行；有序列表序号照常计数
				if t := strings.TrimSpace(mdInlineExceptLists(li)); t != "" {
					sb.WriteString(prefix + t + "\n")
				}
				for sub := li.FirstChild; sub != nil; sub = sub.NextSibling {
					if sub.Type == xhtml.ElementNode && (sub.Data == "ul" || sub.Data == "ol") {
						mdBlocks([]*xhtml.Node{sub}, indent+1, sb)
					}
				}
			}
		case "br":
			// 顶层孤立 br 忽略
		case "div", "p":
			// 容器节点：内部含列表时透明展开逐子节点处理，否则按普通块单行输出
			if hasListNode(n) {
				var cs []*xhtml.Node
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					cs = append(cs, c)
				}
				mdBlocks(cs, indent, sb)
				continue
			}
			emitInlineBlock(n, pad, indent, sb)
		default:
			if richDrop[n.Data] {
				continue
			}
			emitInlineBlock(n, pad, indent, sb)
		}
	}
}

// emitInlineBlock 把节点转为 Markdown 列表项；行内若还嵌着列表则递归展开，避免内容丢失
func emitInlineBlock(n *xhtml.Node, pad string, indent int, sb *strings.Builder) {
	var inline strings.Builder
	var subs []*xhtml.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == xhtml.ElementNode && (c.Data == "ul" || c.Data == "ol") {
			subs = append(subs, c)
			continue
		}
		inline.WriteString(mdInline(c))
	}
	for _, ln := range splitLines(inline.String()) {
		if t := stripListMarker(ln); t != "" {
			sb.WriteString(pad + "- " + t + "\n")
		}
	}
	for _, sub := range subs {
		mdBlocks([]*xhtml.Node{sub}, indent+1, sb)
	}
}

// hasListNode 判断节点的后代中是否存在 ul/ol 列表
func hasListNode(n *xhtml.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != xhtml.ElementNode {
			continue
		}
		if c.Data == "ul" || c.Data == "ol" {
			return true
		}
		if hasListNode(c) {
			return true
		}
	}
	return false
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

// mdToMarkdownStruct 把手工编写的 Markdown 纯文本规整为规范的 Markdown
// （层级缩进、有序标记统一为 "n. "），供导出 .md 时使用。
func mdToMarkdownStruct(s string) string {
	type mline struct {
		level   int
		ordered bool
		num     string
		text    string
	}
	var rows []mline
	for _, raw := range splitLines(s) {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		body := strings.TrimLeft(raw, " \t")
		lead := raw[:len(raw)-len(body)]
		level := len(strings.ReplaceAll(lead, "\t", "  ")) / 2
		ordered := false
		num := ""
		if len(body) >= 2 && (body[0] == '-' || body[0] == '*' || body[0] == '+') && firstRuneIsSpace(body[1:]) {
			body = body[1:]
		} else if k := orderedMarkerLen(body); k > 0 {
			ordered = true
			marker := strings.TrimRight(body[:k], " \t.、)")
			num = marker
			body = body[k:]
		}
		if text := strings.TrimSpace(body); text != "" {
			rows = append(rows, mline{level, ordered, num, text})
		}
	}
	// 与 mdToHTML 相同的层级栈，输出规范缩进
	type frame struct {
		level   int
		ordered bool
	}
	var stack []frame
	closeTop := func() { stack = stack[:len(stack)-1] }
	var sb strings.Builder
	for _, ln := range rows {
		for len(stack) > 0 && stack[len(stack)-1].level > ln.level {
			closeTop()
		}
		if len(stack) > 0 && stack[len(stack)-1].level == ln.level && stack[len(stack)-1].ordered != ln.ordered {
			closeTop()
		}
		if len(stack) == 0 || stack[len(stack)-1].level < ln.level {
			stack = append(stack, frame{level: ln.level, ordered: ln.ordered})
		}
		pad := strings.Repeat("  ", len(stack)-1)
		// 有序列表按行内自带的序号输出，无则从 1 重新计数
		prefix := pad + "- "
		counter := 0
		if ln.ordered {
			counter, _ = strconv.Atoi(ln.num)
			if counter <= 0 {
				counter = 1
			}
			prefix = pad + strconv.Itoa(counter) + ". "
		}
		sb.WriteString(prefix + ln.text + "\n")
	}
	return sb.String()
}
