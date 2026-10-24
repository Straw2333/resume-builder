package service

import (
	"encoding/json"
	"regexp"
	"strings"
)

// RenderMarkdown 将简历渲染为结构化 Markdown 文本，用于导出 .md（同时可直接作为模板重新导入）
func RenderMarkdown(d *RenderData) string {
	var b strings.Builder
	bi := d.BasicInfo
	name := strings.TrimSpace(bi.Name)
	if name == "" {
		name = "未填写姓名"
	}
	b.WriteString("# " + name + "\n\n")
	if bi.Intent != "" {
		b.WriteString("> 求职意向：" + bi.Intent + "\n\n")
	}
	var meta []string
	for _, f := range bi.Fields {
		if f.Visible && strings.TrimSpace(f.Value) != "" {
			meta = append(meta, f.Label+"："+strings.TrimSpace(f.Value))
		}
	}
	if len(meta) > 0 {
		b.WriteString(strings.Join(meta, " ｜ ") + "\n\n")
	}

	for _, s := range d.Sections {
		b.WriteString("## " + s.Title + "\n\n")
		if s.Type == "evaluation" {
			for _, line := range splitLines(s.Content) {
				if line = strings.TrimSpace(line); line != "" {
					b.WriteString("- " + stripListMarker(line) + "\n")
				}
			}
			b.WriteString("\n")
			continue
		}
		for _, it := range s.Items {
			var head []string
			if strings.TrimSpace(it.Title) != "" {
				head = append(head, strings.TrimSpace(it.Title))
			}
			if strings.TrimSpace(it.Subtitle) != "" {
				head = append(head, strings.TrimSpace(it.Subtitle))
			}
			if strings.TrimSpace(it.Time) != "" {
				head = append(head, strings.TrimSpace(it.Time))
			}
			if len(head) > 0 {
				b.WriteString("### " + strings.Join(head, " · ") + "\n\n")
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

// MarkdownTemplate 返回供用户下载填写的 Markdown 简历标准模板
func MarkdownTemplate() string {
	return `<!--
================================================================================
                        简历制作平台 · Markdown 填写模板
================================================================================
填写说明（导入时会自动忽略本注释，无需手动删除）：
1. 一级标题（# ）：填写你的姓名。
2. 引用行（> 求职意向：）：填写目标岗位。
3. 基本信息行：按「字段名：内容」格式填写，多个字段之间用「 ｜ 」或「 | 」分隔，
   也可每行写一条（如 "- 微信：xxx"）。不需要的字段可直接留空或删掉。
4. 二级标题（## ）：模块名称（如：教育背景、实习经历、工作经历、项目经历、专业技能、
   荣誉证书、自我评价，也支持任意自定义模块名称）。不需要的模块可直接整段删除。
5. 三级标题（### ）：模块内的单条经历头部，按「标题 · 副标题 · 时间」用「 · 」或「 | 」分隔：
   - 教育背景：### 学校名称 · 专业 / 学历 · 起止时间
   - 实习/工作经历：### 公司名称 · 职位名称 · 起止时间
   - 项目经历：### 项目名称 · 担任角色 · 起止时间
   - 专业技能：### 技能分类 · 熟练度
   - 荣誉证书：### 证书或奖项名称 · 颁发机构 · 获奖时间
6. 经历描述：使用「- 」无序列表或「1. 」有序列表编写，支持两个空格缩进嵌套子列表，
   重点内容可用 **加粗** 突出显示；导入后平台会自动渲染为精美排版样式。
7. 技能标签（选填）：在条目末尾写「- 技能标签：Go、Redis、Kafka」。
================================================================================
-->

# 张三

> 求职意向：Go 后端开发工程师

性别：男 ｜ 年龄：24 ｜ 所在城市：上海 ｜ 手机号码：13800000000 ｜ 邮箱：zhangsan@example.com ｜ 微信：zhangsan_wx ｜ GitHub：github.com/zhangsan

## 教育背景

### 某某大学 · 计算机科学与技术 / 本科 · 2020.09 - 2024.06

- **主修课程**：数据结构与算法、操作系统、计算机网络、数据库系统原理、分布式系统
- **学业成绩**：GPA 3.8 / 4.0（专业排名前 5%），连续三年获校一等奖学金

## 实习经历

### 某某科技有限公司 · 后端开发实习生 · 2023.06 - 2023.12

- **核心链路优化**：参与交易核心查询接口重构，引入 Redis 多级缓存与 singleflight 防击穿机制，接口 P99 延迟从 **180ms 降至 25ms**，支撑峰值 QPS **2w+**
- **异步任务调度**：基于 Kafka 与延迟队列实现订单超时关单服务，日均稳定处理消息 **500 万+**，线上零故障
- **工程化建设**：完善核心模块单元测试与 Prometheus 监控告警指标，核心代码覆盖率提升至 **85%**
- 技能标签：Go、Gin、GORM、Redis、Kafka

## 项目经历

### 高并发秒杀与优惠券系统 · 核心开发 · 2023.03 - 2023.05

- **防超卖架构**：采用 Redis + Lua 脚本实现库存原子预扣减，结合消息队列异步落库，彻底解决高并发超卖问题
- **限流与降级**：基于令牌桶算法实现网关层接口限流，配合本地缓存兜底热点商品数据，压测单机 QPS 达 **1.5w**
- 技能标签：Go、Redis、Lua、RabbitMQ、MySQL

## 专业技能

### 编程语言 · 熟练

- 熟练掌握 **Go** 语言及其并发编程模型（GMP 调度、Channel、Context），熟悉内存管理与常见性能调优工具（pprof）
- 熟悉 SQL 与 Python，能编写高效的自动化脚本与数据分析工具

### 中间件与数据库 · 掌握

- 熟悉 **MySQL** 索引原理、事务隔离级别与慢查询优化，具备分库分表实践经验
- 熟悉 **Redis** 核心数据结构、持久化机制及缓存穿透/击穿/雪崩解决方案
- 了解 **Kafka**、Docker、Linux 常用命令与 Git 协作流程

## 荣誉证书

### 国家奖学金 · 教育部 · 2023.10

- 本科期间综合测评专业第一名

### 蓝桥杯全国软件和信息技术专业人才大赛国赛二等奖 · 工业和信息化部人才交流中心 · 2022.06

## 自我评价

- 具备扎实的计算机基础知识与良好的编码习惯，崇尚简洁、可维护的工程实践
- 对高并发与分布式系统有浓厚兴趣，习惯通过阅读源码与输出技术博客深入理解底层原理
- 良好的沟通表达与团队协作能力，抗压能力强，能快速融入团队并独立推进需求落地
`
}

var (
	reHTMLComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	reTimeLike    = regexp.MustCompile(`(\d{4}|\d{1,2}\s*[./月年-]\s*\d{1,2}|至今|现在|present|current)`)
)

// 字段中文标签 → 标准 key 与默认图标映射
var labelToFieldMeta = map[string]struct{ key, icon string }{
	"性别":   {"gender", "user"},
	"年龄":   {"age", "calendar"},
	"所在城市": {"location", "map-pin"},
	"城市":   {"location", "map-pin"},
	"现居":   {"location", "map-pin"},
	"现居地":  {"location", "map-pin"},
	"手机号码": {"phone", "phone"},
	"手机":   {"phone", "phone"},
	"电话":   {"phone", "phone"},
	"联系电话": {"phone", "phone"},
	"邮箱":   {"email", "mail"},
	"电子邮箱": {"email", "mail"},
	"Email": {"email", "mail"},
	"政治面貌": {"political", "flag"},
	"籍贯":   {"native", "map-pin"},
	"生日":   {"birthday", "cake"},
	"出生年月": {"birthday", "cake"},
	"学历":   {"degree", "graduation-cap"},
	"最高学历": {"degree", "graduation-cap"},
	"微信":   {"wechat", "message-circle"},
	"微信号":  {"wechat", "message-circle"},
	"GitHub": {"github", "github"},
	"github": {"github", "github"},
	"个人主页": {"homepage", "globe"},
	"主页":   {"homepage", "globe"},
	"博客":   {"homepage", "globe"},
	"工作年限": {"workyears", "briefcase"},
	"工作经验": {"workyears", "briefcase"},
}

// ParseMarkdown 将按模板编写（或平台导出）的 Markdown 文本解析为简历数据结构
func ParseMarkdown(md string) (resumeTitle string, basicJSON json.RawMessage, sectionsJSON json.RawMessage, err error) {
	// 1. 移除 HTML 注释
	clean := reHTMLComment.ReplaceAllString(md, "")
	clean = strings.ReplaceAll(clean, "\r\n", "\n")
	clean = strings.ReplaceAll(clean, "\r", "\n")
	lines := splitLines(clean)

	bi := BasicInfo{}
	var parsedFields []BasicField
	fieldIndex := map[string]int{} // key -> index in parsedFields

	addOrUpdateField := func(label, val string) {
		label = strings.TrimSpace(strings.Trim(label, "*- \t"))
		val = strings.TrimSpace(val)
		if label == "" || val == "" {
			return
		}
		if label == "求职意向" || label == "意向岗位" || label == "目标岗位" {
			if bi.Intent == "" {
				bi.Intent = val
			}
			return
		}
		if label == "姓名" && bi.Name == "" {
			bi.Name = val
			return
		}
		meta, ok := labelToFieldMeta[label]
		if !ok {
			// 大小写不敏感匹配 github / email
			for k, v := range labelToFieldMeta {
				if strings.EqualFold(k, label) {
					meta = v
					ok = true
					break
				}
			}
		}
		key := meta.key
		icon := meta.icon
		if !ok {
			key = "c_" + label
			icon = "none"
		}
		if idx, exists := fieldIndex[key]; exists {
			parsedFields[idx].Value = val
			parsedFields[idx].Visible = true
			return
		}
		fieldIndex[key] = len(parsedFields)
		// 保持标准预设字段名（如“手机”归一为“手机号码”，或保留用户写法均可，预设字段用标准名更统一）
		stdLabel := label
		for _, lf := range legacyFields {
			if lf.key == key {
				stdLabel = lf.label
				break
			}
		}
		parsedFields = append(parsedFields, BasicField{
			Key:     key,
			Label:   stdLabel,
			Value:   val,
			Icon:    icon,
			Visible: true,
		})
	}

	var sections []Section
	var curSec *Section
	var curItem *Item
	var curDescLines []string
	var evalLines []string

	flushItem := func() {
		if curSec == nil || curItem == nil {
			curItem = nil
			curDescLines = nil
			return
		}
		descMD := strings.TrimSpace(strings.Join(curDescLines, "\n"))
		if descMD != "" {
			curItem.Desc = RichHTML(descMD)
		}
		if curItem.Title != "" || curItem.Subtitle != "" || curItem.Time != "" || curItem.Desc != "" || len(curItem.Tags) > 0 {
			curSec.Items = append(curSec.Items, *curItem)
		}
		curItem = nil
		curDescLines = nil
	}

	flushSection := func() {
		if curSec == nil {
			return
		}
		if curSec.Type == "evaluation" {
			curSec.Content = strings.TrimSpace(strings.Join(evalLines, "\n"))
			evalLines = nil
			curSec.Visible = curSec.Content != ""
		} else {
			flushItem()
			curSec.Visible = len(curSec.Items) > 0
		}
		sections = append(sections, *curSec)
		curSec = nil
	}

	inHeader := true
	for _, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" {
			continue
		}
		// 忽略純分隔线 --- / ***
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			continue
		}

		// 二级标题：模块开始（## 模块名）
		if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			flushSection()
			inHeader = false
			secTitle := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			secTitle = strings.Trim(secTitle, "# ")
			if secTitle == "" {
				continue
			}
			curSec = &Section{
				Type:    detectSectionType(secTitle),
				Title:   secTitle,
				Visible: true,
				Items:   []Item{},
			}
			continue
		}

		// 尚在头部（## 出现之前）：解析姓名、求职意向、基本信息字段
		if inHeader {
			if strings.HasPrefix(trimmed, "# ") {
				name := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
				name = strings.Trim(name, "# ")
				if bi.Name == "" {
					bi.Name = name
				}
				continue
			}
			if strings.HasPrefix(trimmed, ">") {
				q := strings.TrimSpace(strings.TrimLeft(trimmed, ">"))
				qParts := splitPipeParts(q)
				for _, qp := range qParts {
					if k, v, ok := splitKV(qp); ok {
						addOrUpdateField(k, v)
					} else if bi.Intent == "" && qp != "" {
						bi.Intent = stripBoldWrap(qp)
					}
				}
				continue
			}
			// 支持同行用 "｜" 或 "|" 分隔的多个字段，或单行 "- 字段：值"
			parts := splitPipeParts(trimmed)
			for _, p := range parts {
				if k, v, ok := splitKV(p); ok {
					addOrUpdateField(k, v)
				}
			}
			continue
		}

		// 模块内部解析
		if curSec == nil {
			continue
		}
		if curSec.Type == "evaluation" {
			// 自我评价模块：忽略可能误写的 ###，直接按行收集
			if strings.HasPrefix(trimmed, "### ") {
				continue
			}
			if line := stripListMarker(trimmed); line != "" {
				evalLines = append(evalLines, line)
			}
			continue
		}

		// 列表型模块：检测三级标题 ### Title · Subtitle · Time
		if strings.HasPrefix(trimmed, "### ") {
			flushItem()
			headLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			t, sub, tm := parseItemHeader(headLine, curSec.Type)
			curItem = &Item{Title: t, Subtitle: sub, Time: tm}
			continue
		}

		// 兼容旧版导出的条目头部格式：以 **标题** 开头且不含列表前缀的行（如 "**公司** · 职位 · 时间"）
		if strings.HasPrefix(trimmed, "**") && (strings.Contains(trimmed, " · ") || strings.HasSuffix(trimmed, "**")) {
			flushItem()
			t, sub, tm := parseItemHeader(trimmed, curSec.Type)
			curItem = &Item{Title: t, Subtitle: sub, Time: tm}
			continue
		}

		// 检查是否为技能标签行（- 技能标签：xxx 或 技能标签：xxx）
		lineNoMarker := stripListMarker(trimmed)
		if isTagLine, tags := parseTagLine(lineNoMarker); isTagLine {
			if curItem == nil {
				curItem = &Item{}
			}
			curItem.Tags = append(curItem.Tags, tags...)
			continue
		}

		// 普通描述行（若用户在模块下未写 ### 直接写描述列表，自动创建一个无标题条目承接）
		if curItem == nil {
			curItem = &Item{}
		}
		curDescLines = append(curDescLines, rawLine)
	}
	flushSection()

	// 组装 BasicInfo：合并默认 5 项基础字段与用户填写的字段
	var finalFields []BasicField
	usedKeys := map[string]bool{}
	for _, lf := range legacyFields {
		if idx, ok := fieldIndex[lf.key]; ok {
			finalFields = append(finalFields, parsedFields[idx])
			usedKeys[lf.key] = true
		} else {
			finalFields = append(finalFields, BasicField{
				Key: lf.key, Label: lf.label, Value: "", Icon: lf.icon, Visible: true,
			})
		}
	}
	for _, pf := range parsedFields {
		if !usedKeys[pf.Key] {
			finalFields = append(finalFields, pf)
		}
	}
	bi.Fields = finalFields
	rawBasic, _ := json.Marshal(bi)
	basicJSON = NormalizeBasic(rawBasic)

	// 补齐未在 Markdown 中出现的预设模块（设为 visible=false，方便用户在编辑器中随时开启）
	defaultSecs := []struct{ typ, title string }{
		{"education", "教育背景"},
		{"skill", "专业技能"},
		{"internship", "实习经历"},
		{"project", "项目经历"},
		{"honor", "荣誉证书"},
		{"evaluation", "自我评价"},
	}
	existingTypes := map[string]bool{}
	for _, s := range sections {
		existingTypes[s.Type] = true
	}
	for _, ds := range defaultSecs {
		if !existingTypes[ds.typ] {
			sections = append(sections, Section{
				Type:    ds.typ,
				Title:   ds.title,
				Visible: false,
				Items:   []Item{},
			})
		}
	}
	sectionsJSON, _ = json.Marshal(sections)

	// 生成默认简历标题：优先「意向-姓名」或「姓名的简历」
	if bi.Name != "" && bi.Intent != "" {
		resumeTitle = bi.Intent + "-" + bi.Name
	} else if bi.Name != "" {
		resumeTitle = bi.Name + "的简历"
	} else {
		resumeTitle = "导入的简历"
	}
	return resumeTitle, basicJSON, sectionsJSON, nil
}

func detectSectionType(title string) string {
	t := strings.ToLower(strings.TrimSpace(title))
	switch {
	case strings.Contains(t, "教育") || strings.Contains(t, "学历") || strings.Contains(t, "学习"):
		return "education"
	case strings.Contains(t, "技能") || strings.Contains(t, "技术") || strings.Contains(t, "能力") || strings.Contains(t, "特长"):
		return "skill"
	case strings.Contains(t, "实习") || strings.Contains(t, "工作") || strings.Contains(t, "任职") || strings.Contains(t, "履历") || strings.Contains(t, "职业"):
		return "internship"
	case strings.Contains(t, "项目") || strings.Contains(t, "作品") || strings.Contains(t, "课题"):
		return "project"
	case strings.Contains(t, "荣誉") || strings.Contains(t, "证书") || strings.Contains(t, "奖项") || strings.Contains(t, "获奖") || strings.Contains(t, "资格"):
		return "honor"
	case strings.Contains(t, "评价") || strings.Contains(t, "总结") || strings.Contains(t, "简介") || strings.Contains(t, "关于"):
		return "evaluation"
	default:
		return "custom"
	}
}

func splitPipeParts(s string) []string {
	normalized := strings.ReplaceAll(s, "｜", "|")
	raw := strings.Split(normalized, "|")
	var out []string
	for _, p := range raw {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitKV(s string) (string, string, bool) {
	s = strings.TrimSpace(strings.TrimLeft(s, "-*+ \t"))
	idx := strings.Index(s, "：")
	sepLen := len("：")
	if idx < 0 {
		idx = strings.Index(s, ":")
		sepLen = 1
	}
	if idx <= 0 {
		return "", "", false
	}
	k := strings.TrimSpace(strings.Trim(s[:idx], "* "))
	v := strings.TrimSpace(strings.TrimLeft(s[idx+sepLen:], "* "))
	v = stripBoldWrap(v)
	if k == "" {
		return "", "", false
	}
	return k, v, true
}

func stripBoldWrap(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 4 && strings.HasPrefix(s, "**") && strings.HasSuffix(s, "**") {
		return strings.TrimSpace(s[2 : len(s)-2])
	}
	return s
}

func parseItemHeader(line, secType string) (title, subtitle, timeStr string) {
	// 支持 " · "、"｜"、"|" 作为分隔符
	normalized := strings.ReplaceAll(line, "｜", " · ")
	normalized = strings.ReplaceAll(normalized, " | ", " · ")
	parts := strings.Split(normalized, "·")
	var clean []string
	for _, p := range parts {
		if c := stripBoldWrap(p); c != "" {
			clean = append(clean, c)
		}
	}
	switch len(clean) {
	case 0:
		return "", "", ""
	case 1:
		return clean[0], "", ""
	case 2:
		if secType != "skill" && reTimeLike.MatchString(clean[1]) {
			return clean[0], "", clean[1]
		}
		return clean[0], clean[1], ""
	default:
		title = clean[0]
		timeStr = clean[len(clean)-1]
		subtitle = strings.Join(clean[1:len(clean)-1], " · ")
		return title, subtitle, timeStr
	}
}

func parseTagLine(s string) (bool, []string) {
	for _, prefix := range []string{"技能标签：", "技能标签:", "标签：", "标签:", "技术栈：", "技术栈:"} {
		if strings.HasPrefix(s, prefix) {
			rest := strings.TrimSpace(strings.TrimPrefix(s, prefix))
			if rest == "" {
				return true, nil
			}
			sep := regexp.MustCompile(`[,，、/｜|\s]+`)
			raw := sep.Split(rest, -1)
			var tags []string
			for _, t := range raw {
				if t = strings.TrimSpace(t); t != "" {
					tags = append(tags, t)
				}
			}
			return true, tags
		}
	}
	return false, nil
}
