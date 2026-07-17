package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type VFS interface {
	SaveFile(filename string, reader io.Reader, overwrite bool) error
	GetFile(filename string) (io.ReadSeekCloser, error)
	DeleteFile(filename string) error
	ListFiles() ([]string, error)
}

type LocalFS struct {
	baseDir string
	locker  *FileLocker
}

func NewLocalFS(dir string) *LocalFS {
	os.MkdirAll(dir, 0755)
	return &LocalFS{
		baseDir: dir,
		locker:  NewFileLocker(),
	}
}

func (fs *LocalFS) SaveFile(filename string, reader io.Reader, overwrite bool) error {
	if !overwrite {
		filename = fs.resolveConflict(filename)
	}

	fs.locker.LockWrite(filename)
	defer fs.locker.UnlockWrite(filename)

	destPath := filepath.Join(fs.baseDir, filename)
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, reader)
	return err
}

func (fs *LocalFS) GetFile(filename string) (io.ReadSeekCloser, error) {
	fs.locker.LockRead(filename)
	
	path := filepath.Join(fs.baseDir, filename)
	file, err := os.Open(path)
	if err != nil {
		fs.locker.UnlockRead(filename)
		return nil, err
	}
	
	return &lockedFile{File: file, locker: fs.locker, filename: filename}, nil
}

func (fs *LocalFS) DeleteFile(filename string) error {
	fs.locker.LockWrite(filename)
	defer fs.locker.UnlockWrite(filename)
	
	path := filepath.Join(fs.baseDir, filename)
	return os.Remove(path)
}

func (fs *LocalFS) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(fs.baseDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func (fs *LocalFS) resolveConflict(filename string) string {
	path := filepath.Join(fs.baseDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return filename
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	for i := 1; ; i++ {
		newFilename := fmt.Sprintf("%s(%d)%s", base, i, ext)
		newPath := filepath.Join(fs.baseDir, newFilename)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newFilename
		}
	}
}

type lockedFile struct {
	*os.File
	locker   *FileLocker
	filename string
}

func (lf *lockedFile) Close() error {
	err := lf.File.Close()
	lf.locker.UnlockRead(lf.filename)
	return err
}
