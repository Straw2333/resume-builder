package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"resume-server/api"
	"resume-server/model"
	"resume-server/service"
)

//go:embed template/resume.html
var resumeTmpl string

// distFS 前端构建产物（构建前由脚本把 web/dist 拷贝到 server/dist）。
// 仓库中仅提交占位文件，未拷贝时编译仍可通过，只是不托管前端。
//
//go:embed all:dist
var distFS embed.FS

func main() {
	if err := os.MkdirAll("data", 0755); err != nil {
		panic("创建数据目录失败: " + err.Error())
	}
	// busy_timeout 缓解并发写时的 database is locked
	db, err := gorm.Open(sqlite.Open("file:data/resume.db?_pragma=busy_timeout(5000)"), &gorm.Config{})
	if err != nil {
		panic("打开数据库失败: " + err.Error())
	}
	if err := db.AutoMigrate(&model.Resume{}); err != nil {
		panic("建表失败: " + err.Error())
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	ra := &api.ResumeAPI{DB: db}
	apiGroup := r.Group("/api/resumes")
	{
		apiGroup.GET("", ra.List)
		apiGroup.POST("", ra.Create)
		apiGroup.GET("/:id", ra.Get)
		apiGroup.PUT("/:id", ra.Update)
		apiGroup.DELETE("/:id", ra.Delete)
		apiGroup.POST("/:id/duplicate", ra.Duplicate)
		apiGroup.GET("/:id/preview", previewHandler(db))
		apiGroup.GET("/:id/export", exportHandler(db))
	}

	// 前端已嵌入时托管静态资源 + SPA 回退（开发模式下 dist 为空则跳过）
	sub, _ := fs.Sub(distFS, "dist")
	indexHTML, indexErr := fs.ReadFile(sub, "index.html")
	hasFrontend := indexErr == nil && len(indexHTML) > 0
	if hasFrontend {
		index := indexHTML
		r.NoRoute(func(c *gin.Context) {
			p := strings.TrimPrefix(c.Request.URL.Path, "/")
			if strings.HasPrefix(p, "api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
				return
			}
			if p == "" {
				p = "index.html"
			}
			if st, err := fs.Stat(sub, p); err == nil && !st.IsDir() {
				if strings.HasPrefix(p, "assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				}
				http.ServeFileFS(c.Writer, c.Request, sub, p)
				return
			}
			// SPA 路由回退
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
		})
	}

	// 端口：8080 被占用时自动换一个可用端口
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			panic("监听端口失败: " + err.Error())
		}
	}
	addr := fmt.Sprintf("http://localhost:%d", ln.Addr().(*net.TCPAddr).Port)
	if hasFrontend {
		fmt.Println("简历平台已启动:", addr, "（按 Ctrl+C 退出）")
		if os.Getenv("NO_OPEN") == "" {
			go func() {
				time.Sleep(300 * time.Millisecond)
				openBrowser(addr)
			}()
		}
	} else {
		fmt.Println("简历平台后端启动:", addr, "（未嵌入前端，开发模式请访问 http://localhost:5173）")
	}
	if err := r.RunListener(ln); err != nil {
		panic(err)
	}
}

// openBrowser 用系统默认浏览器打开页面
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// resolve 载入简历并解析为渲染数据（HTML / Markdown 渲染共用）
func resolve(db *gorm.DB, c *gin.Context) (*model.Resume, *service.RenderData, bool) {
	id, ok := api.ParseID(c)
	if !ok {
		return nil, nil, false
	}
	var m model.Resume
	if err := db.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "简历不存在"})
		return nil, nil, false
	}
	d, err := service.Parse(m.BasicInfo, m.Sections, m.Theme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "简历数据解析失败: " + err.Error()})
		return nil, nil, false
	}
	d.Title = m.Title
	return &m, d, true
}

func previewHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, d, ok := resolve(db, c); ok {
			html, err := service.RenderHTML(resumeTmpl, d)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败: " + err.Error()})
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		}
	}
}

// exportFilename 从简历标题生成安全的下载文件名
func exportFilename(title, ext string) string {
	name := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, strings.TrimSpace(title))
	name = strings.Trim(name, "-")
	if name == "" {
		name = "resume"
	}
	if len(name) > 60 {
		name = name[:60]
	}
	// ASCII 兜底 + RFC 5987 中文文件名
	return fmt.Sprintf(`attachment; filename="resume%s"; filename*=UTF-8''%s%s`,
		ext, url.PathEscape(name), ext)
}

func exportHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		m, d, ok := resolve(db, c)
		if !ok {
			return
		}
		format := c.DefaultQuery("format", "pdf")

		// Markdown：无需临时文件，直接输出
		if format == "md" {
			c.Header("Content-Disposition", exportFilename(m.Title, ".md"))
			c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(service.RenderMarkdown(d)))
			return
		}

		tmpDir, err := os.MkdirTemp("", "resume-*")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer os.RemoveAll(tmpDir)
		htmlPath := filepath.Join(tmpDir, "resume.html")
		html, err := service.RenderHTML(resumeTmpl, d)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败: " + err.Error()})
			return
		}
		if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		switch format {
		case "pdf":
			// 45s 超时，避免浏览器卡死拖住请求
			ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
			defer cancel()
			pdfPath := filepath.Join(tmpDir, "resume.pdf")
			if err := service.ExportPDF(ctx, htmlPath, pdfPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			data, _ := os.ReadFile(pdfPath)
			c.Header("Content-Disposition", exportFilename(m.Title, ".pdf"))
			c.Data(http.StatusOK, "application/pdf", data)
		case "docx":
			// Word 可直接打开并编辑的 HTML 文档
			c.Header("Content-Disposition", exportFilename(m.Title, ".doc"))
			c.Data(http.StatusOK, "application/msword", []byte(html))
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的格式，仅支持 pdf / docx / md"})
		}
	}
}
