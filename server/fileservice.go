package server

import (
	"io"
	"os"
	"path/filepath"
)

// FileService 负责服务端本地的文件系统读写
type FileService struct {
}

type fileMeta interface {
	FilePath() (string, error)
}

func (f *FileService) SaveFile(r io.Reader, fm fileMeta) error {
	filep, errp := fm.FilePath()
	if errp != nil {
		return errp
	}
	if _, err := os.Stat(filepath.Dir(filep)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filep), 0644); err != nil {
			return err
		}
	}
	file, err := os.Create(filep)
	defer file.Close()
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, r); err != nil {
		return err
	}
	return nil
}

func (f *FileService) LoadFile(w io.Writer, fm fileMeta) error {
	filep, errp := fm.FilePath()
	if errp != nil {
		return errp
	}
	file, err := os.Open(filep)
	defer file.Close()
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, file); err != nil {
		return err
	}
	return nil
}
