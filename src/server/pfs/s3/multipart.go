package s3

// code for managing multiparted content

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// TODO: per-uploadID locking
type multipartFileManager struct {
	root            string
	maxAllowedParts int
	lock            *sync.RWMutex
}

func newMultipartFileManager(root string, maxAllowedParts int) *multipartFileManager {
	return &multipartFileManager{
		root:            root,
		maxAllowedParts: maxAllowedParts,
		lock:            &sync.RWMutex{},
	}
}

func (m *multipartFileManager) namePath(uploadID string) string {
	return filepath.Join(m.root, fmt.Sprintf("%s.txt", uploadID))
}

func (m *multipartFileManager) chunksPath(uploadID string) string {
	return filepath.Join(m.root, uploadID)
}

func (m *multipartFileManager) chunkPath(uploadID string, partNumber int) string {
	return filepath.Join(m.chunksPath(uploadID), strconv.Itoa(partNumber))
}

func (m *multipartFileManager) checkExists(uploadID string) error {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if _, err := os.Stat(m.namePath(uploadID)); err != nil {
		return err
	}

	_, err := os.Stat(m.chunksPath(uploadID))
	return err
}

func (m *multipartFileManager) checkChunkExists(uploadID string, partNumber int) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, err := os.Stat(m.chunkPath(uploadID, partNumber))
	return err
}

func (m *multipartFileManager) init(file string) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	uploadID := uuid.NewWithoutDashes()

	if err := ioutil.WriteFile(m.namePath(uploadID), []byte(file), os.ModePerm); err != nil {
		return "", err
	}

	if err := os.Mkdir(m.chunksPath(uploadID), os.ModePerm); err != nil {
		return "", err
	}

	return uploadID, nil
}

func (m *multipartFileManager) filepath(uploadID string) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	name, err := ioutil.ReadFile(m.namePath(uploadID))
	if err != nil {
		return "", err
	}
	return string(name), nil
}

func (m *multipartFileManager) writeChunk(uploadID string, partNumber int, reader io.Reader) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	chunkPath := m.chunkPath(uploadID, partNumber)
	f, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = io.Copy(f, reader); err != nil {
		return err
	}
	return f.Sync()
}

func (m *multipartFileManager) removeChunk(uploadID string, partNumber int) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return os.Remove(m.chunkPath(uploadID, partNumber))
}

func (m *multipartFileManager) listChunks(uploadID string) ([]os.FileInfo, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fileInfos, err := ioutil.ReadDir(m.chunksPath(uploadID))
	if err != nil {
		return nil, err
	}

	// ensure no invalid files exist
	for _, fileInfo := range fileInfos {
		i, err := strconv.Atoi(fileInfo.Name())

		if err != nil || i < 1 || i > m.maxAllowedParts {
			return nil, fmt.Errorf("invalid file exists for %s: %s", uploadID, fileInfo.Name())
		}
	}

	// sort the files
	sort.Slice(fileInfos, func(i, j int) bool {
		// ignore errors since we already verified them
		first, _ := strconv.Atoi(fileInfos[i].Name())
		second, _ := strconv.Atoi(fileInfos[j].Name())
		return first < second
	})

	return fileInfos, nil
}

func (m *multipartFileManager) remove(uploadID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	err := os.Remove(m.namePath(uploadID))
	if err != nil {
		return err
	}
	return os.RemoveAll(m.chunksPath(uploadID))
}

// multipartReader is a reader for multiparted content
type multipartReader struct {
	manager     *multipartFileManager
	uploadID    string
	partNumbers []int
	cur         *os.File
}

func newMultipartReader(manager *multipartFileManager, uploadID string, partNumbers []int) *multipartReader {
	manager.lock.RLock()

	return &multipartReader{
		manager:     manager,
		uploadID:    uploadID,
		partNumbers: partNumbers,
		cur:         nil,
	}
}

func (r *multipartReader) Read(p []byte) (n int, err error) {
	if r.cur == nil {
		if len(r.partNumbers) == 0 {
			return 0, io.EOF
		} else {
			f, err := os.Open(r.manager.chunkPath(r.uploadID, r.partNumbers[0]))
			if err != nil {
				return 0, err
			}
			r.partNumbers = r.partNumbers[1:]
			r.cur = f
		}
	}

	n, err = r.cur.Read(p)
	if err == io.EOF {
		if closeErr := r.cur.Close(); closeErr != nil {
			r.cur = nil
			return n, closeErr
		} else {
			// do not return an EOF, as there may be another chunk to read
			r.cur = nil
			return n, nil
		}
	}
	return n, err
}

func (r *multipartReader) Close() error {
	r.manager.lock.RUnlock()
	if r.cur != nil {
		return r.cur.Close()
	}
	return nil
}
