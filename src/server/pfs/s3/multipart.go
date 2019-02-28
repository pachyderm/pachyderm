package s3

// code for managing multiparted content

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/sirupsen/logrus"
)

// this is a var instead of a const so that we can make a pointer to it
var defaultMaxParts int = 1000

const maxAllowedParts = 10000

// InitiateMultipartUploadResult is an XML-encodable response to initiate a
// new multipart upload
type InitiateMultipartUploadResult struct {
	Bucket   string `xml:"Bucket"`
	Key      string `xml:"Key"`
	UploadID string `xml:"UploadId"`
}

// ListPartsResult is an XML-encodable listing of parts associated with a
// multipart upload
type ListPartsResult struct {
	Bucket               string `xml:"Bucket"`
	Key                  string `xml:"Key"`
	UploadID             string `xml:"UploadId"`
	Initiator            User   `xml:"Initiator"`
	Owner                User   `xml:"Owner"`
	StorageClass         string `xml:"StorageClass"`
	PartNumberMarker     int    `xml:"PartNumberMarker"`
	NextPartNumberMarker int    `xml:"NextPartNumberMarker"`
	MaxParts             int    `xml:"PartNumberMarker"`
	IsTruncated          bool   `xml:"IsTruncated"`
	Part                 []Part `xml:"Part"`
}

func (r *ListPartsResult) isFull() bool {
	return len(r.Part) >= r.MaxParts
}

// CompleteMultipartUpload is an XML-encodable listing of parts associated
// with a multipart upload to complete
type CompleteMultipartUpload struct {
	Parts []Part `xml:"Part"`
}

// Part is an XML-encodable chunk of content associated with a multipart
// upload
type Part struct {
	PartNumber   int       `xml:"PartNumber"`
	LastModified time.Time `xml:"LastModified,omitempty"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml"Size,omitempty"`
}

// multipartFileManager manages the underlying files associated with multipart
// content
//
// multipart content is stored in the local filesystem until it's complete,
// then all of the content is flushed to PFS and the content in the local
// filesystem is removed. This means that ingressing data via multipart upload
// is constrained by local filesystem limitations.
type multipartFileManager struct {
	// the parent directory for all of the multipart contnet
	root string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int

	// a lock to the `locks` mapping
	masterLock *sync.Mutex

	// a mapping of uploadIDs -> their associated r/w locks
	locks map[string]*sync.RWMutex
}

func newMultipartFileManager(root string, maxAllowedParts int) *multipartFileManager {
	return &multipartFileManager{
		root:            root,
		maxAllowedParts: maxAllowedParts,
		masterLock:      &sync.Mutex{},
		locks:           map[string]*sync.RWMutex{},
	}
}

// namePath returns the path to the file storing the filename that will be
// placed in PFS
func (m *multipartFileManager) namePath(uploadID string) string {
	return filepath.Join(m.root, fmt.Sprintf("%s.txt", uploadID))
}

// chunksPath returns the path to the directory storing the chunks/parts
func (m *multipartFileManager) chunksPath(uploadID string) string {
	return filepath.Join(m.root, uploadID)
}

// chunkPath returns the path to the file storing an individual chunk/part
func (m *multipartFileManager) chunkPath(uploadID string, partNumber int) string {
	return filepath.Join(m.chunksPath(uploadID), strconv.Itoa(partNumber))
}

// lock returns the lock associated with an uploadID
func (m *multipartFileManager) lock(uploadID string) *sync.RWMutex {
	m.masterLock.Lock()
	defer m.masterLock.Unlock()

	lock, ok := m.locks[uploadID]
	if !ok {
		lock = &sync.RWMutex{}
		m.locks[uploadID] = lock
	}
	return lock
}

// checkExists checks if an uploadID exists (both the name file and the chunks
// directory)
func (m *multipartFileManager) checkExists(uploadID string) error {
	lock := m.lock(uploadID)
	lock.RLock()
	defer lock.RUnlock()

	if _, err := os.Stat(m.namePath(uploadID)); err != nil {
		return err
	}

	_, err := os.Stat(m.chunksPath(uploadID))
	return err
}

// checkChunkExists checks if a chunk exists
func (m *multipartFileManager) checkChunkExists(uploadID string, partNumber int) error {
	lock := m.lock(uploadID)
	lock.RLock()
	defer lock.RUnlock()
	_, err := os.Stat(m.chunkPath(uploadID, partNumber))
	return err
}

// init starts a new multipart upload, returning its uploadID
func (m *multipartFileManager) init(file string) (string, error) {
	uploadID := uuid.NewWithoutDashes()

	if err := ioutil.WriteFile(m.namePath(uploadID), []byte(file), os.ModePerm); err != nil {
		return "", err
	}

	if err := os.Mkdir(m.chunksPath(uploadID), os.ModePerm); err != nil {
		return "", err
	}

	return uploadID, nil
}

// filepath returns the PFS filepath
func (m *multipartFileManager) filepath(uploadID string) (string, error) {
	lock := m.lock(uploadID)
	lock.RLock()
	defer lock.RUnlock()

	name, err := ioutil.ReadFile(m.namePath(uploadID))
	if err != nil {
		return "", err
	}
	return string(name), nil
}

// writeChunk writes a chunk/part to the local filesystem from a reader
func (m *multipartFileManager) writeChunk(uploadID string, partNumber int, reader io.Reader) error {
	lock := m.lock(uploadID)
	lock.Lock()
	defer lock.Unlock()

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

// removeChunk removes a chunk/part
func (m *multipartFileManager) removeChunk(uploadID string, partNumber int) error {
	lock := m.lock(uploadID)
	lock.Lock()
	defer lock.Unlock()

	return os.Remove(m.chunkPath(uploadID, partNumber))
}

// listChunks lists chunks/parts that have been stored in the local filesystem.
// Returned file infos are sorted by the part number they're associated with.
func (m *multipartFileManager) listChunks(uploadID string) ([]os.FileInfo, error) {
	lock := m.lock(uploadID)
	lock.RLock()
	defer lock.RUnlock()

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

// remove removes an uploadID and all of its content stored in the local filesystem
func (m *multipartFileManager) remove(uploadID string) error {
	lock := m.lock(uploadID)
	lock.Lock()

	err := os.Remove(m.namePath(uploadID))
	if err != nil {
		lock.Unlock()
		return err
	}
	err = os.RemoveAll(m.chunksPath(uploadID))
	if err != nil {
		lock.Unlock()
		return err
	}
	lock.Unlock()
	m.masterLock.Lock()
	defer m.masterLock.Unlock()
	delete(m.locks, uploadID)
	return nil
}

// multipartReader is a reader for multiparted content
type multipartReader struct {
	manager     *multipartFileManager
	uploadID    string
	partNumbers []int

	// the current chunk/part being read
	cur *os.File
}

func newMultipartReader(manager *multipartFileManager, uploadID string, partNumbers []int) *multipartReader {
	lock := manager.lock(uploadID)
	lock.RLock()

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
	lock := r.manager.lock(r.uploadID)
	lock.RUnlock()

	if r.cur != nil {
		return r.cur.Close()
	}
	return nil
}

type multipartHandler struct {
	pc               *client.APIClient
	multipartManager *multipartFileManager
}

func newMultipartHandler(pc *client.APIClient, multipartDir string) *multipartHandler {
	return &multipartHandler{
		pc:               pc,
		multipartManager: newMultipartFileManager(multipartDir, maxAllowedParts),
	}
}

func (h *multipartHandler) init(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	uploadID, err := h.multipartManager.init(file)
	if err != nil {
		writeServerError(w, err)
		return
	}

	result := InitiateMultipartUploadResult{
		Bucket:   branchInfo.Branch.Repo.Name,
		Key:      fmt.Sprintf("%s/%s", branchInfo.Branch.Name, file),
		UploadID: uploadID,
	}

	writeXML(w, http.StatusOK, &result)
}

func (h *multipartHandler) list(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	marker, err := intFormValue(r, "part-number-marker", 1, defaultMaxParts, &defaultMaxParts)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	maxParts, err := intFormValue(r, "max-parts", 1, defaultMaxParts, &defaultMaxParts)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	fileInfos, err := h.multipartManager.listChunks(uploadID)
	if err != nil {
		writeServerError(w, err)
		return
	}

	result := ListPartsResult{
		Bucket:           branchInfo.Branch.Repo.Name,
		Key:              file,
		UploadID:         uploadID,
		Initiator:        defaultUser,
		Owner:            defaultUser,
		StorageClass:     storageClass,
		PartNumberMarker: marker,
		MaxParts:         maxParts,
		IsTruncated:      false,
	}

	for _, fileInfo := range fileInfos {
		// ignore errors converting the name since it's already been verified
		name, _ := strconv.Atoi(fileInfo.Name())

		if name < marker {
			continue
		}
		if result.isFull() {
			result.IsTruncated = true
			break
		}
		result.Part = append(result.Part, Part{
			PartNumber:   name,
			LastModified: fileInfo.ModTime(),
			Size:         fileInfo.Size(),
		})
	}

	if len(result.Part) > 0 {
		result.NextPartNumberMarker = result.Part[len(result.Part)-1].PartNumber
	}

	writeXML(w, http.StatusOK, &result)
}

func (h *multipartHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(r)
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	partNumber, err := intFormValue(r, "partNumber", 1, maxAllowedParts, nil)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	success, err := withBodyReader(r, func(reader io.Reader) bool {
		if err := h.multipartManager.writeChunk(uploadID, partNumber, reader); err != nil {
			writeServerError(w, err)
			return false
		}
		return true
	})

	if err != nil || !success {
		// try to clean up the file if something failed
		removeErr := h.multipartManager.removeChunk(uploadID, partNumber)
		if removeErr != nil {
			logrus.Errorf("could not remove uploadID=%s, partNumber=%d: %v", uploadID, partNumber, removeErr)
		}

		if err != nil {
			writeBadRequest(w, err)
		}

		// if there's no error but the operation is not successful, we've
		// already written a response to the client
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *multipartHandler) complete(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	name, err := h.multipartManager.filepath(uploadID)
	if err != nil {
		writeServerError(w, err)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeServerError(w, fmt.Errorf("could not read request body: %v", err))
		return
	}

	payload := CompleteMultipartUpload{}
	err = xml.Unmarshal(bodyBytes, &payload)
	if err != nil {
		writeBadRequest(w, fmt.Errorf("body is invalid: %v", err))
		return
	}

	// verify that there's at least part, and all parts are in ascending order
	if len(payload.Parts) == 0 {
		writeBadRequest(w, fmt.Errorf("no parts specified"))
		return
	}
	isSorted := sort.SliceIsSorted(payload.Parts, func(i, j int) bool {
		return payload.Parts[i].PartNumber < payload.Parts[j].PartNumber
	})
	if !isSorted {
		writeBadRequest(w, fmt.Errorf("parts not in ascending order"))
		return
	}

	// ensure all the files exist
	for _, part := range payload.Parts {
		if err = h.multipartManager.checkChunkExists(uploadID, part.PartNumber); err != nil {
			writeBadRequest(w, fmt.Errorf("missing part %d", part.PartNumber))
			return
		}
	}

	// pull out the list of part numbers
	partNumbers := []int{}
	for _, part := range payload.Parts {
		partNumbers = append(partNumbers, part.PartNumber)
	}

	// A reader that reads each file chunk. Because this acquires a lock,
	// `Close` MUST be called, or the multipart manager will deadlock
	reader := newMultipartReader(h.multipartManager, uploadID, partNumbers)

	_, err = h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, name, reader, 0)
	if err != nil {
		if closeErr := reader.Close(); closeErr != nil {
			logrus.Errorf("could not close reader for uploadID=%s: %v", uploadID, closeErr)
		}

		writeServerError(w, err)
		return
	}

	if err = reader.Close(); err != nil {
		writeServerError(w, err)
		return
	}

	if err = h.multipartManager.remove(uploadID); err != nil {
		writeServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *multipartHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(r)
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	if err := h.multipartManager.remove(uploadID); err != nil {
		writeServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
