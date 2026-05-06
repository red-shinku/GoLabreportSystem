package service

import (
	"LabSystem/domain"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// FileService 负责服务端本地的文件系统读写
// 需要使用文件业务元数据
type FileService struct {
}

type fileMeta interface {
	FilePath() (string, error)
}

type directoryMeta interface {
	DirectoryPath() (string, error)
}

type pathGenerator interface {
	Next() string
	HasNext() bool
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

func (f *FileService) LoadFile(w io.Writer, path string) error {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return err
	}
	if _, err := io.Copy(w, file); err != nil {
		return err
	}
	return nil
}

// LoadFileBatch 批量加载文件（打包压缩）
func (f *FileService) LoadFileBatch(w io.Writer, fg pathGenerator) {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for fg.HasNext() {
		if err := addFileToZip(zipWriter, fg.Next()); err != nil {
			//FIXME: 后续完善日志
			log.Printf("Error adding file to zip: %v", err)
		}
	}
}

// addFileToZip 将本地文件添加到 zip.Writer 中
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	zipHeader, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}

	safeName := filepath.Base(filePath)
	zipHeader.Name = safeName

	zipHeader.Modified = stat.ModTime()

	// 为文件名启用 UTF-8 编码以避免乱码
	zipHeader.Flags |= 0x800

	writer, err := zipWriter.CreateHeader(zipHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// DeleteFile 删除单文件
func (f *FileService) DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("DeleteFile():%w, %v", domain.ErrNotExist, err)
		}
		return fmt.Errorf("DeleteFile(): %v", err)
	}
	return nil
}

// DeleteDirectory 删除目录
func (f *FileService) DeleteDirectory(fm directoryMeta) error {
	path, err := fm.DirectoryPath()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}
