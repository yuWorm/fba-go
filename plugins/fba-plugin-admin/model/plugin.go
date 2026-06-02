package model

type Plugin struct {
	ID          string   `gorm:"column:id;primaryKey;size:64"`
	Summary     string   `gorm:"column:summary;size:128"`
	Version     string   `gorm:"column:version;size:64"`
	Description string   `gorm:"column:description;type:text"`
	Author      string   `gorm:"column:author;size:64"`
	Tags        []string `gorm:"column:tags;serializer:json;type:json"`
	Database    []string `gorm:"column:database;serializer:json;type:json"`
	DependsOn   []string `gorm:"column:depends_on;serializer:json;type:json"`
	Enabled     bool     `gorm:"column:enabled"`
	BuiltIn     bool     `gorm:"column:built_in"`
}

func (Plugin) TableName() string {
	return "sys_plugin"
}
