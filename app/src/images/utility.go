package images

import (
	"bytes"
	"compress/gzip"
	"image"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"strconv"
	"time"

	"github.com/KzmnbrS/golang_test/src/errors"
	"github.com/disintegration/imaging"
)

// Loads image file from the given multipart form header.
func (im *ImageFile) LoadFromForm(fh *multipart.FileHeader) error {
	fd, err := fh.Open()
	if err != nil {
		return errors.NewIOError(`open`, fh.Filename, err)
	}

	defer fd.Close()

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		return errors.NewIOError(`read`, fh.Filename, err)
	}

	config, extension, err := image.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		return MalformedImage
	}

	im.Contents = contents
	im.Basename = fh.Filename
	im.Extension = extension
	im.Width = config.Width
	im.Height = config.Height
	return nil
}

func UniqueBasename() string {
	prefix := strconv.FormatInt(time.Now().UnixNano(), 36)
	suffix := strconv.FormatInt(rand.Int63n(1024), 36)
	return prefix + suffix
}

// Writes `contents` to the `filepath` and `filepath + ".gz"` (compressed).
// Removes both files on failure. Files are created with 0755 permissions mask.
func WriteFileAndGzip(contents []byte, filepath string) error {
	if err := ioutil.WriteFile(filepath, contents, 0755); err != nil {
		return errors.NewIOError(`write`, filepath, err)
	}

	gzpath := filepath + `.gz`
	gzfile, err := os.OpenFile(gzpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		os.Remove(filepath)
		return errors.NewIOError(`open`, gzpath, err)
	}

	defer gzfile.Close()

	gzwriter := gzip.NewWriter(gzfile)
	defer gzwriter.Close()

	if _, err := gzwriter.Write(contents); err != nil {
		os.Remove(filepath)
		os.Remove(gzpath)
		return errors.NewIOError(`write`, gzpath, err)
	}

	return nil
}

// Loads the parent file from a local drive and generates its preview.
func generatePreview(parentPath string, preview *Preview) (*ImageFile, error) {
	parentImage, err := imaging.Open(parentPath)
	if err != nil {
		return nil, errors.NewIOError(`open`, parentPath, err)
	}

	previewImage := imaging.Fit(
		parentImage,
		preview.Width, preview.Height,
		StringToResampling[preview.Resampling],
	)

	buf := new(bytes.Buffer)
	parentFormat, _ := imaging.FormatFromFilename(parentPath)
	if err := imaging.Encode(buf, previewImage, parentFormat); err != nil {
		return nil, PreviewGenerationFailed
	}

	return &ImageFile{
		Contents: buf.Bytes(),
		Width:    previewImage.Bounds().Dx(),
		Height:   previewImage.Bounds().Dy(),
	}, nil
}
