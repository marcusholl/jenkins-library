package piperutils

import (
	"errors"
	"io"
	"os"
)

//FileUtils ...
type FileUtils struct {
}

// FileExists ...
func (f FileUtils) FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// FileExists ...
func FileExists(filename string) (bool, error) {
	return FileUtils{}.FileExists(filename)
}

// FileCopy ...
func (f FileUtils) FileCopy(src, dst string) (int64, error) {

	exists, err := f.FileExists(src)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, errors.New("Source file '" + src + "' does not exist")
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// FileCopy ...
func FileCopy(src, dst string) (int64, error) {
	return FileUtils{}.FileCopy(src, dst)
}

// FileDelete ...
func (f FileUtils) FileDelete(path string) error {
	return os.Remove(path)
}
