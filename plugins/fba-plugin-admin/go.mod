module github.com/yuWorm/fba-plugin-admin

go 1.25.0

require (
	github.com/yuWorm/fba-go v0.0.0
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.26.1
)

replace github.com/yuWorm/fba-go => ../..
