module github.com/yuWorm/fba-plugin-oauth2

go 1.25.0

require (
	github.com/gofiber/fiber/v3 v3.3.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/yuWorm/fba-go v0.0.0
	github.com/yuWorm/fba-plugin-admin v0.0.0
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.26.1
)

replace github.com/yuWorm/fba-go => ../..

replace github.com/yuWorm/fba-plugin-admin => ../fba-plugin-admin
