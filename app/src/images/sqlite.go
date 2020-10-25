package images

import (
	"os"
	"path"

	"github.com/KzmnbrS/golang_test/src/errors"
	"github.com/jmoiron/sqlx"
)

type SQLiteImageStore struct {
	ImgDir string
	ImgUrl string
	DB     *sqlx.DB
}

func (store *SQLiteImageStore) Push(im *ImageFile) (*Image, error) {
	basename := UniqueBasename() + `.` + im.Extension

	filepath := path.Join(store.ImgDir, basename)
	if err := WriteFileAndGzip(im.Contents, filepath); err != nil {
		return nil, err
	}

	uri := path.Join(store.ImgUrl, basename)
	result := &Image{
		Basename: basename,
		Uri:      uri,
		Width:    im.Width,
		Height:   im.Height,
	}

	id, err := createStoreRecord(store, result)
	if err != nil {
		os.Remove(filepath)
		os.Remove(filepath + `.gz`)
		return nil, err
	}

	result.Id = id
	return result, nil
}

func (store *SQLiteImageStore) List() ([]*Image, error) {
	var images []*Image
	if err := store.DB.Select(&images, `SELECT * FROM image`); err != nil {
		return nil, errors.NewIOError(`select`, `sqlite`, err)
	}

	return images, nil
}

func (store *SQLiteImageStore) Delete(key uint64) error {
	var image Image
	err := store.DB.Get(&image, `SELECT * FROM image WHERE id = $1`, key)
	if err != nil {
		return ImageNotFound
	}

	var relatedPreviews []*Image
	if image.Parent == nil {
		if err := store.DB.Select(
			&relatedPreviews,
			`SELECT * FROM image WHERE parent = $1`, image.Id,
		); err != nil {
			return errors.NewIOError(`select`, `sqlite`, err)
		}
	}

	// Deletion will cascade on related previews
	_, err = store.DB.Exec(`DELETE FROM image WHERE id = $1`, image.Id)
	if err != nil {
		return errors.NewIOError(`delete`, `sqlite`, err)
	}

	for _, preview := range relatedPreviews {
		filepath := path.Join(store.ImgDir, preview.Basename)
		os.Remove(filepath)
		os.Remove(filepath + `.gz`)
	}

	filepath := path.Join(store.ImgDir, image.Basename)
	os.Remove(filepath)
	os.Remove(filepath + `.gz`)
	return nil
}

func (store *SQLiteImageStore) Generate(preview *Preview) (*Image, error) {
	_, resamplingAvailable := StringToResampling[preview.Resampling]
	if !resamplingAvailable {
		return nil, UnsupportedResampling
	}

	parent, err := loadParent(store, preview)
	if err != nil {
		return nil, err
	}

	parentPath := path.Join(store.ImgDir, parent.Basename)
	previewFile, err := generatePreview(parentPath, preview)
	if err != nil {
		return nil, err
	}

	basename := UniqueBasename() + path.Ext(parent.Basename)
	filepath := path.Join(store.ImgDir, basename)
	if err := WriteFileAndGzip(previewFile.Contents, filepath); err != nil {
		return nil, err
	}

	result := &Image{
		Parent:   &parent.Id,
		Basename: basename,
		Uri:      path.Join(store.ImgUrl, basename),
		Width:    previewFile.Width,
		Height:   previewFile.Height,
	}

	id, err := createStoreRecord(store, result)
	if err != nil {
		os.Remove(filepath)
		os.Remove(filepath + `.gz`)
		return nil, err
	}

	result.Id = id
	return result, nil
}

// Creates a new `store` record based on the given `image`, returning its id.
func createStoreRecord(store *SQLiteImageStore, image *Image) (uint64, error) {
	_, err := store.DB.Exec(
		// TODO: RETURNING support?
		`INSERT INTO image VALUES (NULL, $1, $2, $3, $4, $5)`,
		image.Parent, image.Basename, image.Uri, image.Width, image.Height,
	)

	if err != nil {
		return 0, errors.NewIOError(`insert`, `sqlite`, err)
	}

	var id uint64
	err = store.DB.Get(&id, `SELECT id FROM image WHERE uri = $1`, image.Uri)
	if err != nil {
		return 0, errors.NewIOError(`select`, `sqlite`, err)
	}

	return id, nil
}

func loadParent(store *SQLiteImageStore, preview *Preview) (*Image, error) {
	if preview.Width < 28 || preview.Height < 28 {
		return nil, IrrationalPreview
	}

	var parent Image
	if err := store.DB.Get(&parent,
		`SELECT * FROM image WHERE id = $1`, preview.Parent); err != nil {
		// TODO: Or the `store.DB` is not reachable?
		return nil, ImageNotFound
	}

	if preview.Width > parent.Width || preview.Height > parent.Height {
		return nil, IrrationalPreview
	}

	return &parent, nil
}
