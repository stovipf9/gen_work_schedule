package main

import (
	"errors"
	"io"
	"os"
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func makeCacheDir(cacheDirIF CacheDirInterface) error {
	cacheDir, err := CacheDir(cacheDirIF)
	if err != nil {
		return err
	}
	if !fileExists(cacheDir) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}
	}

	return nil
}
