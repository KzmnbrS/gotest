package main

import (
	"log"

	"github.com/KzmnbrS/golang_test/src/images"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func main() {
	setupConfig()
	db := getDatabase(viper.GetString(`DB_PATH`))
	ensurePath(viper.GetString(`IMG_DIR`), true)

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group(`/api`)
	v1 := api.Group(`/v1`)

	images.SetupRoutes(&images.SQLiteImageStore{
		DB:     db,
		ImgDir: viper.GetString(`IMG_DIR`),
		ImgUrl: viper.GetString(`IMG_URL`),
	}, v1.Group(`/images`))

	log.Fatal(app.Listen(`:3000`))
}
