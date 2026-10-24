package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"resume-server/model"
	"resume-server/service"
)

type ResumeAPI struct{ DB *gorm.DB }

// ParseID 校验并解析路径参数 :id，失败时直接写出 400 响应。
// 必须先用它校验，避免把原始字符串交给 GORM 拼进 SQL。
func ParseID(c *gin.Context) (uint, bool) {
	id, ok := parseIDParam(c)
	return id, ok
}

func parseIDParam(c *gin.Context) (uint, bool) {
	id := 0
	if _, err := fmt.Sscan(c.Param("id"), &id); err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return 0, false
	}
	return uint(id), true
}

func (a *ResumeAPI) List(c *gin.Context) {
	var list []model.Resume
	if err := a.DB.Order("updated_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (a *ResumeAPI) Create(c *gin.Context) {
	var r model.Resume
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if r.Title == "" {
		r.Title = "未命名简历"
	}
	if len(r.BasicInfo) == 0 {
		r.BasicInfo = model.DefaultBasicInfo()
	}
	if len(r.Sections) == 0 {
		r.Sections = model.DefaultSections()
	}
	r.BasicInfo = service.NormalizeBasic(r.BasicInfo)
	r.Theme = service.NormalizeTheme(r.Theme)
	if err := a.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (a *ResumeAPI) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var r model.Resume
	if err := a.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "简历不存在"})
		return
	}
	// 归一化旧数据结构，编辑器始终拿到字段列表 + 完整样式配置
	r.BasicInfo = service.NormalizeBasic(r.BasicInfo)
	r.Theme = service.NormalizeTheme(r.Theme)
	c.JSON(http.StatusOK, r)
}

func (a *ResumeAPI) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var r model.Resume
	if err := a.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "简历不存在"})
		return
	}
	// 指针区分"未传字段"与"传了空值"，避免误清空
	var in struct {
		Title     string           `json:"title"`
		Template  string           `json:"template"`
		BasicInfo *json.RawMessage `json:"basicInfo"`
		Sections  *json.RawMessage `json:"sections"`
		Theme     *json.RawMessage `json:"theme"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Title != "" {
		r.Title = in.Title
	}
	if in.Template != "" {
		r.Template = in.Template
	}
	if in.BasicInfo != nil {
		r.BasicInfo = service.NormalizeBasic(*in.BasicInfo)
	}
	if in.Sections != nil {
		r.Sections = *in.Sections
	}
	if in.Theme != nil {
		r.Theme = service.NormalizeTheme(*in.Theme)
	}
	if err := a.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (a *ResumeAPI) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := a.DB.Delete(&model.Resume{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (a *ResumeAPI) Duplicate(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var r model.Resume
	if err := a.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "简历不存在"})
		return
	}
	cp := model.Resume{
		Title:     r.Title + " - 副本",
		Template:  r.Template,
		BasicInfo: r.BasicInfo,
		Sections:  r.Sections,
		Theme:     r.Theme,
	}
	if err := a.DB.Create(&cp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cp)
}

// DownloadTemplate 下载标准 Markdown 简历填写模板
func (a *ResumeAPI) DownloadTemplate(c *gin.Context) {
	c.Header("Content-Disposition", `attachment; filename="resume-template.md"; filename*=UTF-8''%E7%AE%80%E5%8E%86%E5%A1%AB%E5%86%99%E6%A8%A1%E6%9D%BF.md`)
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(service.MarkdownTemplate()))
}

// ImportMarkdown 导入按模板编写的 Markdown 简历（resumeId > 0 时覆盖指定简历，否则新建简历）
func (a *ResumeAPI) ImportMarkdown(c *gin.Context) {
	var in struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		ResumeID uint   `json:"resumeId"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
		return
	}
	if strings.TrimSpace(in.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导入的模板内容不能为空"})
		return
	}
	parsedTitle, basicJSON, sectionsJSON, err := service.ParseMarkdown(in.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模板解析失败: " + err.Error()})
		return
	}

	if in.ResumeID > 0 {
		var r model.Resume
		if err := a.DB.First(&r, in.ResumeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "简历不存在"})
			return
		}
		if r.Title == "" || r.Title == "未命名简历" || r.Title == "我的简历" {
			r.Title = parsedTitle
		}
		r.BasicInfo = basicJSON
		r.Sections = sectionsJSON
		r.Theme = service.NormalizeTheme(r.Theme)
		if err := a.DB.Save(&r).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, r)
		return
	}

	title := strings.TrimSpace(in.Title)
	if title == "" || title == "简历填写模板" || title == "resume-template" {
		title = parsedTitle
	}
	r := model.Resume{
		Title:     title,
		Template:  "classic",
		BasicInfo: basicJSON,
		Sections:  sectionsJSON,
		Theme:     service.NormalizeTheme(model.DefaultTheme()),
	}
	if err := a.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

