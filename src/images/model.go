package images

import (
	"errors"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
)

type Image struct {
	Id uint64 `json:"id"`
	// Parent MIGHT be nil
	Parent *uint64 `json:"parent"`

	Basename string `json:"basename"`
	Uri      string `json:"uri"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

const MAX_IMAGE_SIZE = 134217728 // 128Mb
const IMAGES_DIR = `/opt/persist/images`

type Preview struct {
	Parent     uint64 `json:"parent"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resampling string `json:"resampling"`
}

var StringToResampling = map[string]imaging.ResampleFilter{
	`lanczos`: imaging.Lanczos,
	`catmull`: imaging.CatmullRom,
	`mitnet`:  imaging.MitchellNetravali,
	`linear`:  imaging.Linear,
	`box`:     imaging.Box,
	`nearest`: imaging.NearestNeighbor,
}

var StringToFormat = map[string]imaging.Format{
	`jpeg`: imaging.JPEG,
	`png`:  imaging.PNG,
	`gif`:  imaging.GIF,
}

type ImageStore interface {
	// Saves the image into the `IMAGES_DIR` under a new unique name, preserving
	// its extension. Creates the corresponded `image` store record.
	// FormImage MUST be either jpeg, png or gif.
	Push(*multipart.FileHeader) (Image, error)
	// Generates the `Preview.Parent` image preview based on the given width,
	// height and resampling method.
	GeneratePreview(Preview) (Image, error)
	// Returns a list of all images.
	List() ([]Image, error)
	// Deletes the `key` image from both database and local drive.
	Delete(key uint64) error
}

func uniqueBasename(extension string) string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "." + extension
}

var (
	MalformedImage          = errors.New(`malformed image`)
	ImageIsTooBig           = errors.New(`image is too big`)
	ImageNotFound           = errors.New(`image is not found`)
	IrrationalPreview       = errors.New(`irrational preview`)
	UnsupportedResampling   = errors.New(`unsupported resampling`)
	PreviewGenerationFailed = errors.New(`preview generation failed`)
)

type IOError struct {
	Operation string
	Target    string
	Details   error
}

func (e *IOError) Error() string {
	return e.Operation + " " + e.Target + ": " + e.Details.Error()
}
