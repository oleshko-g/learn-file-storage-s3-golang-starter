package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func saveFileToDisk(src io.Reader, path, ext string) (filePath string, err error) {
	randomData := make([]byte, 32)
	rand.Read(randomData)
	filePath = filepath.Join(path, base64.RawURLEncoding.EncodeToString(randomData)) + "." + ext

	fileOnDisk, err := os.Create(filePath)
	if err == nil {
		defer fileOnDisk.Close()
		_, err = io.Copy(fileOnDisk, src)
	}

	return filePath, err
}
