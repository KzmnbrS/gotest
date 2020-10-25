package images

import (
	"github.com/KzmnbrS/golang_test/src/errors"
	"github.com/disintegration/imaging"
)

type ImageFile struct {
	Contents  []byte
	Basename  string
	Extension string
	Width     int
	Height    int
}

type Image struct {
	Id uint64 `json:"id"`
	// Parent MIGHT be nil
	Parent *uint64 `json:"parent"`

	Basename string `json:"basename"`
	Uri      string `json:"uri"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

type Preview struct {
	Parent     uint64 `json:"parent"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resampling string `json:"resampling"`
}

type ImageStore interface {
	// Saves the image file on the local drive under a new unique name along
	// with its .gz version. Creates a corresponded `Image` store record.
	Push(*ImageFile) (*Image, error)
	// Generates the `Preview.Parent` image preview based on the given width,
	// height and resampling method.
	Generate(*Preview) (*Image, error)
	// Returns a list of all images.
	List() ([]*Image, error)
	// Deletes the `key` image from both database and local drive. If some
	// related previews has been found, they will be deleted either.
	Delete(key uint64) error
}

var StringToResampling = map[string]imaging.ResampleFilter{
	`lanczos`: imaging.Lanczos,
	`catmull`: imaging.CatmullRom,
	`mitnet`:  imaging.MitchellNetravali,
	`linear`:  imaging.Linear,
	`box`:     imaging.Box,
	`nearest`: imaging.NearestNeighbor,
}

var (
	MalformedImage          = errors.NewValueError(`malformed image`)
	ImageIsTooBig           = errors.NewValueError(`image is too big`)
	IrrationalPreview       = errors.NewValueError(`irrational preview`)
	UnsupportedResampling   = errors.NewValueError(`unsupported resampling`)
	PreviewGenerationFailed = errors.NewValueError(`preview generation failed`)
	ImageNotFound           = errors.NewValueError(`image not found`)
)
