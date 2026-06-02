module github.com/yuWorm/fba-go/examples/compat-host

go 1.25.0

require (
	github.com/yuWorm/fba-go v0.0.0
	github.com/yuWorm/fba-plugin-admin v0.0.0
	github.com/yuWorm/fba-plugin-dict v0.0.0
	github.com/yuWorm/fba-plugin-notice v0.0.0
	github.com/yuWorm/fba-plugin-task v0.0.0
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.26.1
)

replace github.com/yuWorm/fba-go => ../..

replace github.com/yuWorm/fba-plugin-admin => ../../plugins/fba-plugin-admin

replace github.com/yuWorm/fba-plugin-dict => ../../plugins/fba-plugin-dict

replace github.com/yuWorm/fba-plugin-notice => ../../plugins/fba-plugin-notice

replace github.com/yuWorm/fba-plugin-task => ../../plugins/fba-plugin-task
