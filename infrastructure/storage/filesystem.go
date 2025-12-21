package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FilesystemDriver реализует Driver в виде файлов в операционной системе
type FilesystemDriver struct {
	rootDir string
	logger  *log.Logger
}

func NewFilesystemDriver(rootDir string, logger *log.Logger) *FilesystemDriver {
	return &FilesystemDriver{
		rootDir: rootDir,
		logger:  logger,
	}
}

func (s *FilesystemDriver) IsLinkedExists(
	ctx context.Context,
	filename string,
	sourceFilename string,
) (exists bool, err error) {
	return s.Exists(ctx, getFullFileName(filename, sourceFilename))
}

func (s *FilesystemDriver) Exists(_ context.Context, filename string) (bool, error) {
	_, err := os.Stat(getFilePath(filename, s.rootDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *FilesystemDriver) UploadLinked(
	ctx context.Context,
	filename string,
	sourceFilename string,
	file []byte,
) error {
	fullFileName := getFullFileName(filename, sourceFilename)
	if exists, err := s.Exists(ctx, fullFileName); exists || err != nil {
		if err != nil {
			return err
		}

		return os.ErrExist
	}
	filePath := getFilePath(fullFileName, s.rootDir)
	err := os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	// Записываем файл
	fstream, err := os.OpenFile(filePath, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fstream.Close()
	_, err = io.Copy(fstream, bytes.NewReader(file))
	if err != nil {
		return err
	}

	return nil
}

func (s *FilesystemDriver) Upload(ctx context.Context, filename string, file []byte) error {
	if exists, err := s.Exists(ctx, filename); exists || err != nil {
		if err != nil {
			return err
		}

		return os.ErrExist
	}
	filePath := getFilePath(filename, s.rootDir)
	err := os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	// Записываем файл
	fstream, err := os.OpenFile(filePath, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fstream.Close()
	_, err = io.Copy(fstream, bytes.NewReader(file))
	if err != nil {
		return err
	}

	return nil
}

func (s *FilesystemDriver) Get(ctx context.Context, filename string) ([]byte, error) {
	if exists, err := s.Exists(ctx, filename); !exists || err != nil {
		if err != nil {
			return nil, err
		}

		return nil, os.ErrNotExist
	}
	filePath := getFilePath(filename, s.rootDir)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (s *FilesystemDriver) GetLinked(ctx context.Context, filename string, sourceFilename string) ([]byte, error) {
	return s.Get(ctx, getFullFileName(filename, sourceFilename))
}

func (s *FilesystemDriver) Delete(_ context.Context, filename string) error {
	// удаляем все связанные картинки
	sourceFilename := strings.Replace(path.Base(filename), path.Ext(filename), "", -1)
	err := os.RemoveAll(getFilePath(path.Join(s.rootDir, sourceFilename), s.rootDir))
	if err != nil {
		return err
	}
	// удаляем исходный файл
	return os.Remove(getFilePath(filename, s.rootDir))
}

func (s *FilesystemDriver) DeleteCache(_ context.Context, filename string, sourceFilename string) error {
	fillFileName := getFullFileName(filename, sourceFilename)
	fillPath := getFilePath(fillFileName, s.rootDir)
	return os.Remove(fillPath)
}

func (s *FilesystemDriver) GetSpaceUsage(_ context.Context) (usage int64, err error) {
	usage = int64(0)
	err = filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			usage += info.Size()
		}
		return nil
	})

	return usage, err
}
