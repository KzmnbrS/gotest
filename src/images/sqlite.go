package images

import (
	"bytes"
	"compress/gzip"
	"image"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"

	"github.com/disintegration/imaging"
	"github.com/jmoiron/sqlx"
)

type SQLiteImageStore struct {
	*sqlx.DB
	ImagesUrl string
}

func (store *SQLiteImageStore) Push(fh *multipart.FileHeader) (Image, error) {
	if fh.Size > MAX_IMAGE_SIZE {
		return Image{}, ImageIsTooBig
	}

	fd, err := fh.Open()
	if err != nil {
		return Image{}, &IOError{
			Operation: `open`,
			Target:    fh.Filename,
			Details:   err,
		}
	}

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		return Image{}, &IOError{
			Operation: `read`,
			Target:    fh.Filename,
			Details:   err,
		}
	}

	config, extension, err := image.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		return Image{}, MalformedImage
	}

	basename := uniqueBasename(extension)
	uri := path.Join(store.ImagesUrl, basename)

	filepath := path.Join(IMAGES_DIR, basename)
	gzfilepath := filepath + `.gz`
	if err := writeFileAndGzip(contents, filepath, gzfilepath); err != nil {
		return Image{}, err
	}

	result, err := createFileRecord(store, Image{
		Basename: basename,
		Uri:      uri,
		Width:    config.Width,
		Height:   config.Height,
	})

	if err != nil {
		os.Remove(filepath)
		os.Remove(gzfilepath)
		return Image{}, err
	}

	return result, nil
}

func (store *SQLiteImageStore) List() ([]Image, error) {
	var images []Image
	if err := store.Select(&images, `SELECT * FROM image`); err != nil {
		return nil, &IOError{
			Operation: `read`,
			Target:    `sqlite`,
			Details:   err,
		}
	}

	return images, nil
}

func (store *SQLiteImageStore) Delete(key uint64) error {
	var image Image
	err := store.Get(&image, `SELECT * FROM image WHERE id = $1`, key)
	if err != nil {
		return ImageNotFound
	}

	// Deletion will cascade on image previews
	_, err = store.Exec(`DELETE FROM image WHERE id = $1`, image.Id)
	if err != nil {
		return &IOError{
			Operation: `delete`,
			Target:    `sqlite`,
			Details:   err,
		}
	}

	isParent := image.Parent == nil
	if isParent {
		var previewList []Image
		err = store.Select(
			&previewList,
			`SELECT * FROM image WHERE parent = $1`,
			image.Id,
		)

		for _, preview := range previewList {
			filepath := path.Join(IMAGES_DIR, preview.Basename)
			os.Remove(filepath)
			os.Remove(filepath + `.gz`)
		}
	}

	filepath := path.Join(IMAGES_DIR, image.Basename)
	os.Remove(filepath)
	os.Remove(filepath + `.gz`)
	return nil
}

func (store *SQLiteImageStore) GeneratePreview(preview Preview) (Image, error) {
	resampling, resamplingAvailable := StringToResampling[preview.Resampling]
	if !resamplingAvailable {
		return Image{}, UnsupportedResampling
	}

	var parent Image
	if err := store.Get(
		&parent,
		`SELECT * FROM image WHERE id = $1`,
		preview.Parent,
	); err != nil {
		return Image{}, ImageNotFound
	}

	if preview.Width > parent.Width || preview.Height > parent.Height {
		return Image{}, IrrationalPreview
	}

	if preview.Width < 28 || preview.Height < 28 {
		return Image{}, IrrationalPreview
	}

	parentPath := path.Join(IMAGES_DIR, parent.Basename)
	parentImage, err := imaging.Open(parentPath)
	if err != nil {
		return Image{}, &IOError{
			Operation: `open`,
			Target:    parentPath,
			Details:   err,
		}
	}

	parentExtension := path.Ext(parent.Basename)[1:]
	parentFormat := StringToFormat[parentExtension]

	previewImage := imaging.Fit(
		parentImage,
		preview.Width, preview.Height,
		resampling,
	)

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, previewImage, parentFormat); err != nil {
		return Image{}, PreviewGenerationFailed
	}

	basename := uniqueBasename(parentExtension)
	uri := path.Join(store.ImagesUrl, basename)

	filepath := path.Join(IMAGES_DIR, basename)
	gzfilepath := filepath + `.gz`
	if err := writeFileAndGzip(buf.Bytes(), filepath, gzfilepath); err != nil {
		return Image{}, err
	}

	result, err := createFileRecord(store, Image{
		Parent:   &parent.Id,
		Basename: basename,
		Uri:      uri,
		Width:    previewImage.Bounds().Dx(),
		Height:   previewImage.Bounds().Dy(),
	})

	if err != nil {
		os.Remove(filepath)
		os.Remove(gzfilepath)
		return Image{}, err
	}

	return result, nil
}

// Writes `contents` to the `filepath` and `gzfilepath` (compressed). Removes
// both files on failure. Files are created with O_TRUNC mode and 0755 mask.
func writeFileAndGzip(contents []byte, filepath, gzfilepath string) error {
	if err := ioutil.WriteFile(filepath, contents, 0755); err != nil {
		return &IOError{
			Operation: `write`,
			Target:    filepath,
			Details:   err,
		}
	}

	gzfile, err := os.OpenFile(
		gzfilepath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0755,
	)

	if err != nil {
		os.Remove(filepath)
		return &IOError{
			Operation: `open`,
			Target:    gzfilepath,
			Details:   err,
		}
	}

	defer gzfile.Close()

	gzwriter := gzip.NewWriter(gzfile)
	defer gzwriter.Close()

	if _, err := gzwriter.Write(contents); err != nil {
		os.Remove(filepath)
		os.Remove(gzfilepath)
		return &IOError{
			Operation: `write`,
			Target:    gzfilepath,
			Details:   err,
		}
	}

	return nil
}

// Creates a new `store` record based on the given `image`. Returns the same
// `image` with a new id.
func createFileRecord(store *SQLiteImageStore, image Image) (Image, error) {
	_, err := store.Exec(
		// TODO: RETURNING support?
		`INSERT INTO image VALUES (NULL, $1, $2, $3, $4, $5)`,
		image.Parent, image.Basename, image.Uri, image.Width, image.Height,
	)

	if err != nil {
		return Image{}, &IOError{
			Operation: `insert`,
			Target:    `sqlite`,
			Details:   err,
		}
	}

	store.Get(&image, `SELECT * FROM image WHERE uri = $1`, image.Uri)
	return image, nil
}
