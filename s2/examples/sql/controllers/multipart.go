package controllers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/s2"
	"github.com/pachyderm/s2/examples/sql/models"
	"github.com/pachyderm/s2/examples/sql/util"
)

func (c *Controller) ListMultipart(r *http.Request, name, keyMarker, uploadIDMarker string, maxUploads int) (*s2.ListMultipartResult, error) {
	c.logger.Tracef("ListMultipart: name=%+v, keyMarker=%+v, uploadIDMarker=%+v, maxUploads=%+v", name, keyMarker, uploadIDMarker, maxUploads)

	result := s2.ListMultipartResult{
		Uploads: []*s2.Upload{},
	}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		uploads, err := models.ListUploads(tx, bucket.ID, keyMarker, uploadIDMarker, maxUploads+1)
		if err != nil {
			return err
		}

		for _, upload := range uploads {
			if len(result.Uploads) >= maxUploads {
				if maxUploads > 0 {
					result.IsTruncated = true
				}
				break
			}

			result.Uploads = append(result.Uploads, &s2.Upload{
				Key:          upload.Key,
				UploadID:     upload.ID,
				Initiator:    models.GlobalUser,
				StorageClass: models.StorageClass,
				Initiated:    models.Epoch,
			})
		}

		return nil
	})

	return &result, err
}

func (c *Controller) InitMultipart(r *http.Request, name, key string) (string, error) {
	c.logger.Tracef("InitMultipart: name=%+v, key=%+v", name, key)

	result := ""

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		upload, err := models.CreateUpload(tx, bucket.ID, key)
		if err != nil {
			return err
		}

		result = upload.ID
		return nil
	})

	return result, err
}

func (c *Controller) AbortMultipart(r *http.Request, name, key, uploadID string) error {
	c.logger.Tracef("AbortMultipart: name=%+v, key=%+v, uploadID=%+v", name, key, uploadID)

	return c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		_, err = models.GetUpload(tx, bucket.ID, key, uploadID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchUploadError(r)
			}
			return err
		}

		return models.DeleteUpload(tx, bucket.ID, key, uploadID)
	})
}

func (c *Controller) CompleteMultipart(r *http.Request, name, key, uploadID string, parts []*s2.Part) (*s2.CompleteMultipartResult, error) {
	c.logger.Tracef("CompleteMultipart: name=%+v, key=%+v, uploadID=%+v, parts=%+v", name, key, uploadID, parts)

	result := s2.CompleteMultipartResult{
		Location: models.Location,
	}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		_, err = models.GetUpload(tx, bucket.ID, key, uploadID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchUploadError(r)
			}
			return err
		}

		content := []byte{}

		for i, part := range parts {
			uploadPart, err := models.GetUploadPart(tx, uploadID, part.PartNumber)
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					return s2.InvalidPartError(r)
				}
				return err
			}
			if fmt.Sprintf("\"%s\"", uploadPart.ETag) != part.ETag {
				return s2.InvalidPartError(r)
			}
			// each part, except for the last, is expected to be at least 5mb in
			// s3
			if i < len(parts)-1 && len(uploadPart.Content) < 5*1024*1024 {
				return s2.EntityTooSmallError(r)
			}

			content = append(content, uploadPart.Content...)
		}

		version := "null"
		if bucket.Versioning == s2.VersioningEnabled {
			version = util.RandomString(10)
			result.Version = version
		}

		object, err := models.CreateObjectContent(tx, bucket.ID, key, version, content)
		if err != nil {
			return err
		}

		result.ETag = object.ETag

		if err := models.DeleteUpload(tx, bucket.ID, key, uploadID); err != nil {
			return err
		}

		return nil
	})

	return &result, err
}

func (c *Controller) ListMultipartChunks(r *http.Request, name, key, uploadID string, partNumberMarker, maxParts int) (*s2.ListMultipartChunksResult, error) {
	c.logger.Tracef("ListMultipartChunks: name=%+v, key=%+v, uploadID=%+v, partNumberMarker=%+v, maxParts=%+v", name, key, uploadID, partNumberMarker, maxParts)

	result := s2.ListMultipartChunksResult{
		Initiator:    &models.GlobalUser,
		Owner:        &models.GlobalUser,
		StorageClass: models.StorageClass,
		Parts:        []*s2.Part{},
	}

	err := c.transaction(func(tx *gorm.DB) error {
		_, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		uploadParts, err := models.ListUploadParts(tx, uploadID, partNumberMarker, maxParts+1)
		if err != nil {
			return err
		}

		for _, uploadPart := range uploadParts {
			if len(result.Parts) >= maxParts {
				if maxParts > 0 {
					result.IsTruncated = true
				}
				break
			}

			result.Parts = append(result.Parts, &s2.Part{
				PartNumber: uploadPart.Number,
				ETag:       uploadPart.ETag,
			})
		}

		return nil
	})

	return &result, err
}

func (c *Controller) UploadMultipartChunk(r *http.Request, name, key, uploadID string, partNumber int, reader io.Reader) (string, error) {
	c.logger.Tracef("UploadMultipartChunk: name=%+v, key=%+v, uploadID=%+v partNumber=%+v", name, key, uploadID, partNumber)

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	result := ""

	err = c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		_, err = models.GetUpload(tx, bucket.ID, key, uploadID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchUploadError(r)
			}
			return err
		}

		uploadPart, err := models.UpsertUploadPart(tx, uploadID, partNumber, content)
		if err != nil {
			return err
		}

		result = uploadPart.ETag
		return nil
	})

	return result, err
}
