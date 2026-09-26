package model

// Setting 通用键值配置表。目前存 AI 服务配置（key=ai_settings，value 为 JSON），
// 单机应用无多用户，直接明文存本地 SQLite。
type Setting struct {
	Key   string `gorm:"primarykey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
