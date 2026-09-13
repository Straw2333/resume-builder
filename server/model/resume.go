package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Resume 简历主表。BasicInfo/Sections/Theme 存 JSON，天然支持模块选填、自定义字段与主题配色。
type Resume struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	Title     string          `gorm:"size:128" json:"title"`
	Template  string          `gorm:"size:32;default:classic" json:"template"`
	BasicInfo json.RawMessage `gorm:"type:text" json:"basicInfo"`
	Sections  json.RawMessage `gorm:"type:text" json:"sections"`
	Theme     json.RawMessage `gorm:"type:text" json:"theme"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Section 类型常量：basic 之外的模块均可选填
const (
	SectionEducation  = "education"  // 教育背景
	SectionSkill      = "skill"      // 专业技能
	SectionInternship = "internship" // 实习经历
	SectionProject    = "project"    // 项目经历
	SectionHonor      = "honor"      // 荣誉证书
	SectionEvaluation = "evaluation" // 自我评价
	SectionCustom     = "custom"     // 自定义模块
)

// DefaultBasicInfo 新建简历的基本信息：姓名/求职意向固定，其余为可增删、可开关的字段
func DefaultBasicInfo() json.RawMessage {
	return json.RawMessage(`{
		"name": "",
		"intent": "",
		"avatar": "",
		"fields": [
			{"key":"gender","label":"性别","value":"","icon":"user","visible":true},
			{"key":"age","label":"年龄","value":"","icon":"calendar","visible":true},
			{"key":"location","label":"所在城市","value":"","icon":"map-pin","visible":true},
			{"key":"phone","label":"手机号码","value":"","icon":"phone","visible":true},
			{"key":"email","label":"邮箱","value":"","icon":"mail","visible":true}
		]
	}`)
}

// DefaultTheme 默认主题：蓝黑
func DefaultTheme() json.RawMessage {
	return json.RawMessage(`{"dark":"#16213e","accent":"#2563eb"}`)
}

// DefaultSections 默认模块骨架：全部包含但 visible=false，用户按需开启
func DefaultSections() json.RawMessage {
	return json.RawMessage(`[
		{"type":"education","title":"教育背景","visible":true,"items":[]},
		{"type":"skill","title":"专业技能","visible":true,"items":[]},
		{"type":"internship","title":"实习经历","visible":false,"items":[]},
		{"type":"project","title":"项目经历","visible":true,"items":[]},
		{"type":"honor","title":"荣誉证书","visible":false,"items":[]},
		{"type":"evaluation","title":"自我评价","visible":false,"content":""},
		{"type":"custom","title":"自定义模块","visible":false,"items":[]}
	]`)
}
