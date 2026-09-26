package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"resume-server/model"
	"resume-server/service"
)

const aiSettingsKey = "ai_settings"

// AIAPI AI 优化相关接口：设置存取、连通性测试、简历润色。
type AIAPI struct{ DB *gorm.DB }

// loadSettings 读配置；从未配置时返回零值结构（不报错，由调用方决定提示语）。
func (a *AIAPI) loadSettings() (*service.AISettings, error) {
	var s model.Setting
	err := a.DB.First(&s, "key = ?", aiSettingsKey).Error
	if err == gorm.ErrRecordNotFound {
		return &service.AISettings{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg service.AISettings
	if err := json.Unmarshal([]byte(s.Value), &cfg); err != nil {
		return &service.AISettings{}, nil
	}
	return &cfg, nil
}

func (a *AIAPI) saveSettings(cfg *service.AISettings) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	s := model.Setting{Key: aiSettingsKey, Value: string(b)}
	if err := a.DB.Save(&s).Error; err != nil {
		return err
	}
	return nil
}

// settingsView 返回给前端的视图：Key 不回显明文，只给尾 4 位。
type settingsView struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"baseURL"`
	Model    string `json:"model"`
	HasKey   bool   `json:"hasKey"`
	KeyTail  string `json:"keyTail"`
}

func toView(cfg *service.AISettings) settingsView {
	v := settingsView{Provider: cfg.Provider, BaseURL: cfg.BaseURL, Model: cfg.Model}
	if cfg.APIKey != "" {
		v.HasKey = true
		tail := cfg.APIKey
		if len(tail) > 4 {
			tail = tail[len(tail)-4:]
		}
		v.KeyTail = tail
	}
	return v
}

func (a *AIAPI) GetSettings(c *gin.Context) {
	cfg, err := a.loadSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	c.JSON(http.StatusOK, toView(cfg))
}

// SaveSettings 保存配置；apiKey 留空表示沿用旧值（方便只改模型/地址不重填 Key）。
func (a *AIAPI) SaveSettings(c *gin.Context) {
	old, err := a.loadSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var in struct {
		Provider string `json:"provider"`
		BaseURL  string `json:"baseURL"`
		Model    string `json:"model"`
		APIKey   string `json:"apiKey"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg := &service.AISettings{
		Provider: in.Provider,
		BaseURL:  in.BaseURL,
		Model:    in.Model,
		APIKey:   in.APIKey,
	}
	if cfg.APIKey == "" {
		cfg.APIKey = old.APIKey
	}
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	if err := cfg.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveSettings(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toView(cfg))
}

// Test 用给定配置发一条测试消息。apiKey 为空时沿用已保存的 Key（方便保存前测试）。
func (a *AIAPI) Test(c *gin.Context) {
	old, err := a.loadSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var in service.AISettings
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.APIKey == "" {
		in.APIKey = old.APIKey
	}
	if in.Provider == "" {
		in.Provider = "openai"
	}
	if err := in.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reply, err := service.Chat(&in, "你是连通性测试助手。", "请只回复两个字：正常")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "连接成功，模型回复：" + reply})
}

// Polish 单条润色：内容由前端当前编辑状态传入，本接口不读写简历数据。
func (a *AIAPI) Polish(c *gin.Context) {
	cfg, err := a.loadSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var in service.PolishInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	html, err := service.Polish(cfg, &in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"html": html})
}
