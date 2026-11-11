package security

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func NewPrivateLogger(filename string) (*log.Logger, *os.File, error) {
	if filename == "" || filepath.Base(filename) != filename {
		return nil, nil, errors.New("invalid log filename")
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, nil, err
	}
	appDir := filepath.Join(cacheDir, "buster-client")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return nil, nil, err
	}
	if err := os.Chmod(appDir, 0700); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(appDir, filename)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, nil, errors.New("log target is not a regular file")
	}
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return nil, nil, err
	}
	return log.New(file, "", log.Ldate|log.Ltime|log.LUTC), file, nil
}
