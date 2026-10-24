package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"resume-server/model"
	"resume-server/service"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open memory sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Resume{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	r := gin.New()
	ra := &ResumeAPI{DB: db}
	g := r.Group("/api/resumes")
	{
		g.GET("/template", ra.DownloadTemplate)
		g.POST("/import", ra.ImportMarkdown)
		g.GET("/:id", ra.Get)
	}
	return r, db
}

func TestDownloadAndImportTemplateAPI(t *testing.T) {
	r, db := setupTestRouter(t)

	// 1. 测试下载模板接口
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/resumes/template", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	tmplBody := w.Body.String()
	if !strings.Contains(tmplBody, "# 张三") {
		t.Fatalf("template content missing expected header")
	}

	// 2. 测试导入新建简历接口
	payload, _ := json.Marshal(map[string]any{
		"title":   "简历填写模板",
		"content": tmplBody,
	})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/resumes/import", bytes.NewReader(payload))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("import create failed: %d %s", w2.Code, w2.Body.String())
	}
	var created model.Resume
	if err := json.Unmarshal(w2.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if created.ID == 0 || created.Title != "Go 后端开发工程师-张三" {
		t.Errorf("unexpected created resume: id=%d, title=%s", created.ID, created.Title)
	}

	// 3. 测试在编辑页导入模板覆盖当前简历（保留原自定义配色 Theme）
	customTheme := json.RawMessage(`{"dark":"#4c1d24","accent":"#e05d6f"}`)
	db.Model(&created).Update("theme", service.NormalizeTheme(customTheme))

	overrideMD := "# 赵六\n> 求职意向：云原生工程师\n\n## 项目经历\n### K8s 算子开发 · 负责人 · 2024.01 - 2024.08\n- 编写自定义 Operator"
	payload3, _ := json.Marshal(map[string]any{
		"resumeId": created.ID,
		"content":  overrideMD,
	})
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/resumes/import", bytes.NewReader(payload3))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("import override failed: %d %s", w3.Code, w3.Body.String())
	}
	var updated model.Resume
	_ = json.Unmarshal(w3.Body.Bytes(), &updated)
	if !strings.Contains(string(updated.BasicInfo), "赵六") {
		t.Errorf("expected basicInfo updated to 赵六, got: %s", string(updated.BasicInfo))
	}
	if !strings.Contains(string(updated.Theme), "#4c1d24") {
		t.Errorf("expected custom theme preserved, got: %s", string(updated.Theme))
	}
}
