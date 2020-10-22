package main

import (
	"log"
	"os"
	"strconv"

	"github.com/KzmnbrS/golang_test/src/images"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sqlx.Connect(
		`sqlite3`,
		`/opt/persist/images.db`,
	)

	db.Exec(`PRAGMA foreign_keys = ON`)

	if err != nil {
		log.Fatal(err)
	}

	store := images.SQLiteImageStore{
		DB:        db,
		ImagesUrl: os.Getenv(`IMAGES_URL`),
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	app.Post(`/images`, func(c *fiber.Ctx) error {
		fh, err := c.FormFile(`image`)
		if err != nil {
			return c.SendStatus(400)
		}

		image, err := store.Push(fh)
		if err != nil {
			if _, isIoError := err.(*images.IOError); isIoError {
				return c.SendStatus(500)
			}

			return c.Status(400).JSON(fiber.Map{
				`error`: err.Error(),
			})
		}

		return c.JSON(image)
	})

	app.Post(`/images/:image_id/preview`, func(c *fiber.Ctx) error {
		imageId, err := strconv.ParseUint(c.Params(`image_id`, `foo`), 10, 64)
		if err != nil {
			return c.SendStatus(400)
		}

		var payload images.Preview
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		payload.Parent = imageId
		preview, err := store.GeneratePreview(payload)
		if err != nil {
			if _, isIoError := err.(*images.IOError); isIoError {
				return c.SendStatus(500)
			}

			return c.Status(400).JSON(fiber.Map{
				`error`: err.Error(),
			})
		}

		return c.JSON(preview)
	})

	app.Get(`/images`, func(c *fiber.Ctx) error {
		images, err := store.List()
		if err != nil {
			return c.SendStatus(500)
		}

		return c.JSON(images)
	})

	app.Delete(`/images/:image_id`, func(c *fiber.Ctx) error {
		imageId, err := strconv.ParseUint(c.Params(`image_id`, `foo`), 10, 64)
		if err != nil {
			return c.SendStatus(400)
		}

		err = store.Delete(imageId)
		if err != nil {
			if _, isIoError := err.(*images.IOError); isIoError {
				return c.SendStatus(500)
			}

			return c.Status(400).JSON(fiber.Map{
				`error`: err.Error(),
			})
		}

		return c.SendStatus(200)
	})

	log.Fatal(app.Listen(`:3000`))
}
