# 简历制作平台

一个使用 **Go + Vue 3** 构建的本地简历管理 / 制作 / 导出平台。无需登录、零外部依赖（SQLite 单文件数据库），内置 AI 润色与根据岗位 JD 生成能力，下载简历支持 PDF / Word / Markdown。

![tech](https://img.shields.io/badge/Go-Gin%20%2B%20GORM-00ADD8) ![tech](https://img.shields.io/badge/Vue-3%20%2B%20Element%20Plus-42b883) ![db](https://img.shields.io/badge/DB-SQLite-003B57)

## 功能特性

### 简历管理
- 多份简历管理：新建、编辑、复制、删除、卡片式列表展示
- 自动保存：编辑 1.2 秒后自动保存并刷新预览，支持 **Ctrl+S / ⌘S** 手动保存
- 离开保护：有未保存更改时关闭页面 / 切换路由会弹窗确认

### 模块化简历内容（全部选填）
基本信息固定在最前，其余模块均可开关、排序、增删：

| 模块 | 说明 |
|---|---|
| 基本信息 | 姓名与求职意向固定，其余字段可增删/开关/选图标，支持头像上传 |
| 教育背景 / 专业技能 / 实习经历 / 项目经历 / 荣誉证书 | 多条目，支持拖拽/上下排序 |
| 自我评价 | 多行文本 |
| 自定义模块 | 可添加多个（如「开源经历」、「校园活动」），标题自拟 |

- **富文本描述**：条目描述支持加粗（Ctrl+B）、无序列表、有序列表、增减缩进（Tab / Shift+Tab），粘贴自动去格式；服务端白名单净化防 XSS，PDF/Word/Markdown 导出同步转换（加粗→`**粗体**` 等）。
- **Markdown 快捷输入**：描述框内输入 `- 项目描述`、`1. 内容` 后按空格（或整行输完按回车）即自动转为对应列表；`**重点**` 后按空格自动加粗；粘贴带 `- ` / `1. ` / `**加粗**` 标记的纯文本也会直接结构化。
- **模板导入与下载**：支持一键下载标准 Markdown 填写模板（.md）；支持上传导入外部 Markdown 简历，服务端深度语法解析，既可一键新建简历，也可一键导入覆盖已有简历（自动保留原有主题色彩与排版样式）。

### AI 智能辅助（本地配置 Key，多协议兼容）
- **根据 JD 定向生成简历**：输入目标岗位的招聘 JD，大模型自动提取岗位核心要求与意向，生成量身定制的项目经历与技能点，并支持按模块精准按需填充。
- **整份简历优化与 Diff 报告**：一键并行优化全篇经历条目（遵循 STAR 原则），生成左右可视差异报告（高亮增删改），支持全选或逐条选择采纳；优化前自动备份，支持随时**一键恢复**。
- **单条精准润色与多轮追问**：针对单条项目/工作经历进行深度润色（提供强化数据、精简表达、突出成果等多种预设方向），并支持在弹窗内通过自然语言多轮追问微调。
- **多协议与私有模型支持**：支持 OpenAI / Claude 兼容协议接口，自由配置 API Key、Base URL 与 Model 名称，可无缝对接 DeepSeek、Kimi、通义千问、Claude、ChatGPT 等海内外主流大模型或本地 Ollama。

### 样式定制
- **A4 智能分页预览**：预览与 PDF 导出按标准 A4（794×1123px）分页，内容超过一页时自动「智能合页」——以条目为单位整块换页、不在行中间截断，跨页模块自动补「模块名（续）」标题，PDF 页数与预览完全一致。
- **文字**：中文字体（雅黑/苹方/思源/宋体/黑体/楷体）、英文字体（Arial/Times/Georgia 等）、正文字号（12–18px）、行间距（1.2–2.4）。
- **样式**：头部布局（居中/居左/平铺）、信息展示（图标/文字/纯内容）、编辑面板支持自由拖拽调节宽度。
- **配色**：标题底色 + 强调色取色器，内置蓝黑/墨黑/青碧/酒红预设。
- **信息展示**：预置性别、年龄、政治面貌、籍贯、生日、学历、微信、GitHub、个人主页、工作年限等，支持自定义字段与 15+ 种精美图标切换。

### 导出
| 格式 | 实现方式 | 说明 |
|---|---|---|
| PDF | 本机 Chrome/Edge `--headless --print-to-pdf` | 与预览像素级一致，中文正常 |
| Word | Word 兼容 HTML（.doc） | 打开即可继续编辑 |
| Markdown | 服务端模板渲染 | 加粗/列表/缩进自动转换 |

导出文件名自动使用简历标题（RFC 5987 编码，支持中文）。

## 技术栈

| 端 | 技术 |
|---|---|
| 后端 | Go 1.21+、Gin、GORM、SQLite（[glebarez/sqlite](https://github.com/glebarez/sqlite)，纯 Go 无 CGO）、`go:embed` 内嵌模板 |
| 前端 | Vue 3、Vite 5、Element Plus、Pinia、Vue Router |
| 富文本 | 原生 contenteditable + `document.execCommand`，后端 `x/net/html` 白名单净化 |

## 目录结构

```
Resume/
├── server/                  # Go 后端
│   ├── main.go              # 路由、静态嵌入、自动拉起浏览器
│   ├── model/               # 数据模型（Resume、Setting）
│   ├── api/
│   │   ├── resume.go        # 简历 CRUD、模板下载与导入
│   │   └── ai.go            # AI 配置、单条润色、整份优化、JD 生成
│   ├── service/
│   │   ├── render.go        # 数据归一化与 HTML 渲染
│   │   ├── render_md.go     # 双向 Markdown 解析与模板渲染
│   │   ├── richtext.go      # 富文本白名单净化 + HTML/MD 互转
│   │   ├── llm.go           # OpenAI/Claude 协议客户端封装
│   │   └── export.go        # Chrome/Edge 无头浏览器 PDF 导出
│   ├── template/resume.html # A4 智能分页简历渲染模板
│   └── data/resume.db       # SQLite 本地数据库（运行时生成，不入库）
└── web/                     # Vue 3 前端
    └── src/
        ├── views/           # List（管理列表） / Editor（主编辑台） / Preview（大图预览）
        ├── components/
        │   ├── RichTextEditor.vue      # Markdown 快捷富文本编辑器
        │   ├── AISettingsDialog.vue    # 大模型接口/Key 设置弹窗
        │   ├── AIPolishDialog.vue      # 单条精准润色与追问弹窗
        │   ├── AIOptimizeAllDialog.vue # 整份优化 Diff 比对与采纳弹窗
        │   └── AIJDGenerateDialog.vue  # 基于岗位 JD 生成简历内容弹窗
        ├── utils/
        │   └── aiDiff.js               # 词级/行级差异对比高亮算法
        ├── api/resume.js    # 后端接口请求封装
        └── router/          # 前端路由
```

## 快速开始

### 环境要求
- Go 1.21+（后端）
- Node.js 18+（前端开发）
- 本机装有 Chrome 或 Edge（仅 PDF 导出需要；Word/Markdown 不依赖）

### 开发模式（前后端分离）

```bash
# 终端 1：后端（端口 8080）
cd server
go run .

# 终端 2：前端（端口 5173，已配置代理转发 /api → 8080）
cd web
npm install
npm run dev
```

浏览器访问 **http://localhost:5173**

### 生产部署（单文件发布）

前端构建产物会嵌入二进制（`go:embed`），**一个 exe 即完整程序**，双击运行自动打开浏览器：

```bash
# 一键构建（Windows）
build.bat

# 或手动：构建前端 → 拷贝到 server/dist → 构建后端
cd web && npm run build
cd ../server
rm -rf dist && mkdir dist && cp -r ../web/dist/* dist/   # Windows 用 xcopy
CGO_ENABLED=0 go build -ldflags "-s -w" -o 简历制作平台.exe .
```

- 产物约 19MB，纯 Go 零依赖，拷贝到任意 Windows/Linux/macOS 机器直接运行
- 8080 端口被占用时自动换可用端口；环境变量 `NO_OPEN=1` 可禁用自动开浏览器
- 数据库自动生成在 exe 同目录 `data/` 下，绿色便携

### 下载现成安装包

仓库配置了 GitHub Actions（`.github/workflows/release.yml`）：推送 `v*` 标签即自动构建 Windows / Linux / macOS 三个平台的可执行文件并发布到 [Releases](https://github.com/Straw2333/resume-builder/releases)：

```bash
git tag v1.0.0 && git push origin v1.0.0
```

### Docker（可选示例）

单文件二进制天然适合容器——一个拷贝即可运行：

```dockerfile
FROM node:18-alpine AS web
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS server
WORKDIR /app
COPY server/go.mod server/go.sum ./server/
COPY server/ ./server/
COPY --from=web /app/web/dist ./server/dist
WORKDIR /app/server
RUN CGO_ENABLED=0 go build -o /resume-server .

FROM alpine:3.19
# PDF 导出需要 chromium；仅用 Word/MD 导出可去掉
RUN apk add --no-cache chromium
COPY --from=server /resume-server /usr/local/bin/
EXPOSE 8080
CMD ["resume-server"]
```

> Docker 镜像内浏览器路径与 `service/export.go` 的探测路径不同，需自行挂载或调整 `browserCandidates()`。

## API 一览

| 分组 | 方法 | 路径 | 说明 |
|---|---|---|---|
| **简历基础** | GET / POST | `/api/resumes` | 简历列表 / 新建空简历 |
| | GET / PUT / DELETE | `/api/resumes/:id` | 简历详情 / 自动保存 / 删除 |
| | POST | `/api/resumes/:id/duplicate` | 快速复制简历 |
| | GET | `/api/resumes/:id/preview` | HTML 实时分页渲染预览 |
| | GET | `/api/resumes/:id/export?format=pdf\|docx\|md` | 导出下载（PDF/Word/Markdown） |
| **模板交互** | GET | `/api/resumes/template` | 下载标准 Markdown 简历填写模板 |
| | POST | `/api/resumes/import` | 导入 Markdown 内容（支持覆盖或新建） |
| **图片上传** | POST | `/api/upload` | 头像与图片上传 |
| | GET | `/api/uploads/:name` | 静态图片资源访问 |
| **AI 接口** | GET / PUT | `/api/ai/settings` | 获取 / 保存大模型 API 接口配置 |
| | POST | `/api/ai/test` | 连通性测试 |
| | POST | `/api/ai/polish` | 单条经历内容润色 / 多轮追问微调 |
| | POST | `/api/ai/generate-from-jd` | 根据岗位招聘 JD 定向生成简历内容 |

## 数据模型

单表 `resumes`，内容字段以 JSON 存储，天然支持模块/字段扩展而无需迁移：

```go
type Resume struct {
    ID        uint
    Title     string          // 简历名称
    BasicInfo json.RawMessage // {name, intent, avatar, fields[]}
    Sections  json.RawMessage // [{type, title, visible, items[] | content}]
    Theme     json.RawMessage // {dark, accent, fontCN, fontEN, fontSize,
                              //  lineHeight, headerLayout, fieldStyle}
}
```

读取与保存时服务端自动做**归一化**：旧版扁平基本信息自动迁移为字段列表、非法主题值回落默认，前后端数据结构永远一致。

## 常见问题

- **PDF 导出报"未找到 Chrome 或 Edge"**：`service/export.go` 按常见安装路径探测浏览器，若安装路径特殊可修改 `browserCandidates()`；Word 与 Markdown 导出不依赖浏览器。
- **端口被占用**：后端默认 `:8080`（`main.go` 中 `r.Run(":8080")`），前端代理目标在 `web/vite.config.js`。
- **数据库位置**：`server/data/resume.db`，相对启动目录；备份/迁移直接拷贝该文件即可。

## License

MIT
