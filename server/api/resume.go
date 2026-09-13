package api

import (
	"fmt"
	"net/http"

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
	var in model.Resume
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Title != "" {
		r.Title = in.Title
	}
	r.Template = in.Template
	if len(in.BasicInfo) > 0 {
		r.BasicInfo = service.NormalizeBasic(in.BasicInfo)
	}
	if len(in.Sections) > 0 {
		r.Sections = in.Sections
	}
	if len(in.Theme) > 0 {
		r.Theme = service.NormalizeTheme(in.Theme)
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
