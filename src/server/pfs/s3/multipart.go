package s3

// code for managing multiparted content

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/sirupsen/logrus"
)

const multipartRepo = "_s3gateway_multipart_"

const defaultMaxParts = 1000

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
	MaxParts             int    `xml:"MaxParts"`
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
	Size         uint64    `xml"Size,omitempty"`
}

// multipartFileManager manages the underlying files associated with multipart
// content.
type multipartFileManager struct {
	// Name of the PFS repo holding multipart content
	repo string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int

	pc *client.APIClient
}

func newMultipartFileManager(repo string, maxAllowedParts int, pc *client.APIClient) (*multipartFileManager, error) {
	err := pc.CreateRepo(repo)
	if err != nil && !strings.Contains(err.Error(), "as it already exists") {
		return nil, err
	}

	m := multipartFileManager{
		repo:            repo,
		maxAllowedParts: maxAllowedParts,
		pc:              pc,
	}

	return &m, nil
}

// checkExists checks if an uploadID exists
func (m *multipartFileManager) checkExists(uploadID string) error {
	_, err := m.pc.InspectFile(m.repo, uploadID, "filepath.txt")
	return err
}

// checkChunkExists checks if a chunk exists
func (m *multipartFileManager) checkChunkExists(uploadID string, partNumber int) error {
	_, err := m.pc.InspectFile(m.repo, uploadID, fmt.Sprintf("%d.chunk", partNumber))
	return err
}

// init starts a new multipart upload, returning its uploadID
func (m *multipartFileManager) init(file string) (string, error) {
	// don't use the full UUID because UUIDs are specifically not allowed as
	// branch names
	uploadID := uuid.NewWithoutDashes()[:16]

	err := m.pc.CreateBranch(m.repo, uploadID, "", nil)
	if err != nil {
		return "", err
	}

	// NOTE: if something goes wrong here, the branch isn't deleted and will
	// need to be manually deleted
	_, err = m.pc.PutFileOverwrite(m.repo, uploadID, "filepath.txt", strings.NewReader(file), 0)
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

// filepath returns the PFS filepath
func (m *multipartFileManager) filepath(uploadID string) (string, error) {
	reader, err := m.pc.GetFileReader(m.repo, uploadID, "filepath.txt", 0, 0)
	if err != nil {
		return "", err
	}
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// writeChunk writes a chunk/part from a reader
func (m *multipartFileManager) writeChunk(uploadID string, partNumber int, reader io.Reader) error {
	_, err := m.pc.PutFileOverwrite(m.repo, uploadID, fmt.Sprintf("%d.chunk", partNumber), reader, 0)
	return err
}

// removeChunk removes a chunk/part
func (m *multipartFileManager) removeChunk(uploadID string, partNumber int) error {
	return m.pc.DeleteFile(m.repo, uploadID, fmt.Sprintf("%d.chunk", partNumber))
}

// listChunks lists chunks/parts that have been stored. Returned file infos
// are sorted by the part number they're associated with.
func (m *multipartFileManager) listChunks(uploadID string) ([]Part, error) {
	parts := []Part{}
	fileInfos, err := m.pc.GlobFile(m.repo, uploadID, "*.chunk")
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		path := fileInfo.File.Path[1:]
		i, err := strconv.Atoi(path[:len(path)-6])
		if err != nil || i < 1 || i > m.maxAllowedParts {
			return nil, fmt.Errorf("invalid file exists for %s: %s", uploadID, path)
		}

		timestamp, err := types.TimestampFromProto(fileInfo.Committed)
		if err != nil {
			return nil, err
		}

		parts = append(parts, Part{
			PartNumber:   i,
			LastModified: timestamp,
			Size:         fileInfo.SizeBytes,
		})
	}

	// sort the parts
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	return parts, nil
}

// listChunks lists chunks/parts that have been stored, sorted by part number
func (m *multipartFileManager) complete(uploadID string, partNumbers []int, destRepo string, destBranch string) error {
	destPath, err := m.filepath(uploadID)
	if err != nil {
		return err
	}

	for _, partNumber := range partNumbers {
		srcPath := fmt.Sprintf("%d.chunk", partNumber)
		err = m.pc.CopyFile(m.repo, uploadID, srcPath, destRepo, destBranch, destPath, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// remove removes an uploadID and all of its contents
func (m *multipartFileManager) remove(uploadID string) error {
	return m.pc.DeleteBranch(m.repo, uploadID, false)
}

type multipartHandler struct {
	pc               *client.APIClient
	multipartManager *multipartFileManager
}

func newMultipartHandler(pc *client.APIClient) (*multipartHandler, error) {
	multipartManager, err := newMultipartFileManager(multipartRepo, maxAllowedParts, pc)
	if err != nil {
		return nil, err
	}

	h := multipartHandler{
		pc:               pc,
		multipartManager: multipartManager,
	}

	return &h, nil
}

func (h *multipartHandler) init(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(w, r)

	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	uploadID, err := h.multipartManager.init(file)
	if err != nil {
		internalError(w, r, err)
		return
	}

	result := InitiateMultipartUploadResult{
		Bucket:   fmt.Sprintf("%s.%s", branchInfo.Branch.Name, branchInfo.Branch.Repo.Name),
		Key:      file,
		UploadID: uploadID,
	}

	writeXML(w, http.StatusOK, &result)
}

func (h *multipartHandler) list(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(w, r)
	
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		noSuchUploadError(w, r)
		return
	}

	marker := intFormValue(r, "part-number-marker", 1, defaultMaxParts, 0)
	maxParts := intFormValue(r, "max-parts", 1, defaultMaxParts, defaultMaxParts)

	parts, err := h.multipartManager.listChunks(uploadID)
	if err != nil {
		internalError(w, r, err)
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
	
	for _, part := range parts {
		if part.PartNumber < marker {
			continue
		}
		if result.isFull() {
			result.IsTruncated = true
			break
		}
		result.Part = append(result.Part, part)
	}

	if len(result.Part) > 0 {
		result.NextPartNumberMarker = result.Part[len(result.Part)-1].PartNumber
	}

	writeXML(w, http.StatusOK, &result)
}

func (h *multipartHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(w, r)
	
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		noSuchUploadError(w, r)
		return
	}

	partNumber := intFormValue(r, "partNumber", 1, maxAllowedParts, -1)
	if partNumber == -1 {
		// TODO: it's not clear from s3 docs whether this is the right error
		invalidPartError(w, r)
		return
	}

	success := withBodyReader(w, r, func(reader io.Reader) bool {
		if err := h.multipartManager.writeChunk(uploadID, partNumber, reader); err != nil {
			internalError(w, r, err)
			return false
		}
		return true
	})

	if !success {
		// try to clean up the file if something failed
		err = h.multipartManager.removeChunk(uploadID, partNumber)
		if err != nil {
			logrus.Errorf("could not remove uploadID=%s, partNumber=%d: %v", uploadID, partNumber, err)
		}
	}
}

func (h *multipartHandler) complete(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(w, r)
	
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		noSuchUploadError(w, r)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("could not read request body: %v", err)
		internalError(w, r, err)
		return
	}
	payload := CompleteMultipartUpload{}
	err = xml.Unmarshal(bodyBytes, &payload)
	if err != nil {
		malformedXMLError(w, r)
		return
	}

	// verify that there's at least part, and all parts are in ascending order
	isSorted := sort.SliceIsSorted(payload.Parts, func(i, j int) bool {
		return payload.Parts[i].PartNumber < payload.Parts[j].PartNumber
	})
	if len(payload.Parts) == 0 || !isSorted {
		invalidPartOrderError(w, r)
		return
	}

	// ensure all the files exist
	for _, part := range payload.Parts {
		if err = h.multipartManager.checkChunkExists(uploadID, part.PartNumber); err != nil {
			invalidPartError(w, r)
			return
		}
	}

	// pull out the list of part numbers
	partNumbers := []int{}
	for _, part := range payload.Parts {
		partNumbers = append(partNumbers, part.PartNumber)
	}

	err = h.multipartManager.complete(uploadID, partNumbers, branchInfo.Branch.Repo.Name, branchInfo.Branch.Name)
	if err != nil {
		// TODO: handle missing repo/branch
		internalError(w, r, err)
		return
	}

	if err = h.multipartManager.remove(uploadID); err != nil {
		internalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *multipartHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch, _ := objectArgs(w, r)
	
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	uploadID := r.FormValue("uploadId")
	if err := h.multipartManager.checkExists(uploadID); err != nil {
		noSuchUploadError(w, r)
		return
	}
	if err := h.multipartManager.remove(uploadID); err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
