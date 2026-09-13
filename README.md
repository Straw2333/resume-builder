# 简历制作平台

一个使用 **Go + Vue 3** 构建的本地简历管理 / 制作 / 导出平台。无需登录、零外部依赖（SQLite 单文件数据库），下载简历支持 PDF / Word / Mark。

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
| 基本信息 | 姓名与求职意向固定，其余字段可增删/开关/选图标 |
| 教育背景 / 专业技能 / 实习经历 / 项目经历 / 荣誉证书 | 多条目，支持排序 |
| 自我评价 | 多行文本 |
| 自定义模块 | 可添加多个（如「校园经历」），标题自拟 |

**富文本描述**：条目描述支持加粗（Ctrl+B）、无序列表、有序列表、增减缩进（Tab / Shift+Tab），粘贴自动去格式；服务端白名单净化防 XSS，PDF/Word/Markdown 导出同步转换（加粗→`**粗体**` 等）。

### 样式定制
- **文字**：中文字体（雅黑/苹方/思源/宋体/黑体/楷体）、英文字体（Arial/Times/Georgia 等）、正文字号（12–18px）、行间距（1.2–2.4）
- **样式**：头部布局（居中/居左/平铺）、信息展示（图标/文字/纯内容）
- **配色**：标题底色 + 强调色取色器，内置蓝黑/墨黑/青碧/酒红预设
- 基本信息字段预置：性别、年龄、**政治面貌**、籍贯、生日、学历、微信、GitHub、个人主页、工作年限等，支持自定义字段；15 种 lucide 风格图标可选

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
│   ├── main.go              # 路由、预览/导出接口、模板 embed
│   ├── model/resume.go      # GORM 模型与默认数据
│   ├── api/resume.go        # CRUD Handler（含 ID 校验防注入）
│   ├── service/
│   │   ├── render.go        # 数据解析、归一化、HTML 渲染
│   │   ├── render_md.go     # Markdown 渲染
│   │   ├── richtext.go      # 富文本白名单净化 + MD 转换
│   │   └── export.go        # 无头浏览器 PDF 导出（45s 超时）
│   ├── template/resume.html # 简历纸张模板（预览/PDF/Word 共用）
│   └── data/resume.db       # SQLite 数据库（运行时生成）
└── web/                     # Vue 3 前端
    └── src/
        ├── views/           # List / Editor / Preview
        ├── components/
        │   └── RichTextEditor.vue  # 富文本描述编辑器
        ├── api/resume.js    # axios 封装
        └── router/          # 路由
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

### 生产部署（单进程）

```bash
# 构建前端静态资源
cd web
npm run build            # 产物输出到 web/dist

# 构建后端二进制
cd ../server
go build -o resume-server.exe .   # Linux: go build -o resume-server .
./resume-server.exe              # 启动于 :8080
```

生产模式下前端构建产物由 Vite 输出至 `web/dist`，可通过任意静态服务器托管，或将 `dist` 目录交给 Nginx：API 反代 `/api → 127.0.0.1:8080`。

**Nginx 参考配置**：

```nginx
server {
    listen 80;
    root /var/www/resume/web/dist;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
    }
    location / {
        try_files $uri $uri/ /index.html;   # SPA 路由回退
    }
}
```

### Docker（可选示例）

```dockerfile
FROM node:18-alpine AS web
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS server
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -o /resume-server .

FROM alpine:3.19
# PDF 导出需要 chromium；仅用 Word/MD 导出可去掉
RUN apk add --no-cache chromium
COPY --from=server /resume-server /usr/local/bin/
COPY --from=web /app/web/dist /var/www/html
EXPOSE 8080
CMD ["resume-server"]
```

> Docker 镜像内浏览器路径与 `service/export.go` 的探测路径不同，需自行挂载或调整 `browserCandidates()`。

## API 一览

| 方法 | 路径 | 说明 |
|---|---|---|
| GET / POST | `/api/resumes` | 列表 / 新建 |
| GET / PUT / DELETE | `/api/resumes/:id` | 详情 / 保存 / 删除 |
| POST | `/api/resumes/:id/duplicate` | 复制 |
| GET | `/api/resumes/:id/preview` | HTML 预览 |
| GET | `/api/resumes/:id/export?format=pdf\|docx\|md` | 导出下载 |

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
