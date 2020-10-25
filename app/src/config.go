package main

import (
	"log"
	"os"

	v "github.com/spf13/viper"
)

func setupConfig() {
	v.SetDefault(`IMG_MAX_SIZE`, 134217728 /* 128Mb */)
	v.BindEnv(`IMG_MAX_SIZE`)

	v.SetDefault(`IMG_URL`, `/static`)
	v.BindEnv(`IMG_URL`)

	v.SetDefault(`IMG_DIR`, `/opt/persist`)
	v.BindEnv(`IMG_DIR`)

	v.SetDefault(`DB_PATH`, `/opt/persist/images.db`)
	v.BindEnv(`DB_PATH`)
}

func ensurePath(path string, isDirectory bool) bool {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf(`Path ensuring error: %v`, err)
		}

		if isDirectory {
			if err := os.Mkdir(path, 0666); err != nil {
				log.Fatalf(`Can't create a directory: %v`, err)
			}
		} else {
			file, err := os.Create(path)
			if err != nil {
				log.Fatalf(`Can't create a file: %v`, err)
			}

			file.Close()
		}

		return true
	}

	return false
}
