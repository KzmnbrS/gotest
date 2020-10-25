package images

import (
	"net/http"
	"strconv"

	"github.com/KzmnbrS/golang_test/src/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func SetupRoutes(store ImageStore, route fiber.Router) {
	route.Post(``, postImages(store, viper.GetInt64(`IMG_MAX_SIZE`)))
	route.Post(`/:image_id/preview`, postImages_imageId_Preview(store))
	route.Get(``, getImages(store))
	route.Delete(`/:image_id`, deleteImages_imageId(store))
}

func postImages(store ImageStore, imageMaxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fh, err := c.FormFile(`image`)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		if fh.Size > imageMaxSize {
			return c.SendStatus(http.StatusRequestEntityTooLarge)
		}

		file := ImageFile{}
		if err := file.LoadFromForm(fh); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		image, err := store.Push(&file)
		if err != nil {
			return errortext(err, c)
		}

		return c.JSON(image)
	}
}

func postImages_imageId_Preview(store ImageStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		imageId, err := parseImageId(`image_id`, c)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		var payload Preview
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		payload.Parent = imageId
		preview, err := store.Generate(&payload)
		if err != nil {
			return errortext(err, c)
		}

		return c.JSON(preview)
	}
}

func getImages(store ImageStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		images, err := store.List()
		if err != nil {
			return errortext(err, c)
		}

		if images == nil {
			return c.JSON([]int{})
		}

		return c.JSON(images)
	}
}

func deleteImages_imageId(store ImageStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		imageId, err := parseImageId(`image_id`, c)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		if err := store.Delete(imageId); err != nil {
			return errortext(err, c)
		}

		return c.SendStatus(http.StatusOK)
	}
}

func errortext(e error, c *fiber.Ctx) error {
	switch e.(type) {
	case *errors.IOError:
		return c.SendStatus(http.StatusInternalServerError)

	case *errors.ValueError:
		{
			var status int
			if e == ImageNotFound {
				status = http.StatusNotFound
			} else {
				status = http.StatusBadRequest
			}

			return c.Status(status).JSON(fiber.Map{`error`: e.Error()})
		}

	default:
		return c.SendStatus(http.StatusInternalServerError)
	}
}

func parseImageId(key string, c *fiber.Ctx) (uint64, error) {
	return strconv.ParseUint(c.Params(key, `foo`), 10, 64)
}
