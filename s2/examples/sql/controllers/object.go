package controllers

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/s2"
	"github.com/pachyderm/s2/examples/sql/models"
	"github.com/pachyderm/s2/examples/sql/util"
)

func (c *Controller) GetObject(r *http.Request, name, key, version string) (*s2.GetObjectResult, error) {
	c.logger.Tracef("GetObject: name=%+v, key=%+v, version=%+v", name, key, version)

	result := s2.GetObjectResult{
		ModTime: models.Epoch,
	}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		var object models.Object
		if bucket.Versioning == s2.VersioningEnabled && version != "" {
			object, err = models.GetObject(tx, bucket.ID, key, version)
		} else {
			object, err = models.GetLatestObject(tx, bucket.ID, key)
		}
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchKeyError(r)
			}
			return err
		}

		if bucket.Versioning == s2.VersioningEnabled {
			result.Version = object.Version
		}

		if object.DeleteMarker {
			if bucket.Versioning == s2.VersioningEnabled {
				result.DeleteMarker = true
			} else {
				return s2.NoSuchKeyError(r)
			}
		} else {
			result.ETag = object.ETag
			result.Content = bytes.NewReader(object.Content)
		}

		return nil
	})

	return &result, err
}

func (c *Controller) CopyObject(r *http.Request, srcBucket, srcKey string, obj *s2.GetObjectResult, destBucket, destKey string) (string, error) {
	c.logger.Tracef("CopyObject: srcBucket=%+v, srcKey=%+v, obj=%+v, destBucket=%+v, destKey=%+v", srcBucket, srcKey, obj, destBucket, destKey)
	version, _, err := c.putObject(r, destBucket, destKey, obj.Content)
	return version, err
}

func (c *Controller) PutObject(r *http.Request, name, key string, reader io.Reader) (*s2.PutObjectResult, error) {
	c.logger.Tracef("PutObject: name=%+v, key=%+v", name, key)
	version, etag, err := c.putObject(r, name, key, reader)
	if err != nil {
		return nil, err
	}
	return &s2.PutObjectResult{
		Version: version,
		ETag:    etag,
	}, nil
}

func (c *Controller) DeleteObject(r *http.Request, name, key, version string) (*s2.DeleteObjectResult, error) {
	c.logger.Tracef("DeleteObject: name=%+v, key=%+v, version=%+v", name, key, version)

	result := s2.DeleteObjectResult{}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		var object models.Object
		if version != "" && bucket.Versioning == s2.VersioningEnabled {
			object, err = models.GetObject(tx, bucket.ID, key, version)
		} else {
			object, err = models.GetLatestObject(tx, bucket.ID, key)
		}
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			return err
		}

		if gorm.IsRecordNotFoundError(err) {
			object, err = models.CreateObjectDeleteMarker(tx, bucket.ID, key, util.RandomString(10))
			if err != nil {
				return err
			}

			result.Version = object.Version
		} else if object.DeleteMarker {
			if err = tx.Delete(&object).Error; err != nil {
				return err
			}

			result.DeleteMarker = true
		} else {
			object.DeleteMarker = true
			object.ETag = ""
			object.Content = nil
			if err = tx.Save(&object).Error; err != nil {
				return err
			}

			result.Version = object.Version
		}

		return nil
	})

	return &result, err
}

func (c *Controller) putObject(r *http.Request, name, key string, reader io.Reader) (string, string, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", "", err
	}

	version := ""
	etag := ""

	err = c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		if bucket.Versioning == s2.VersioningEnabled {
			object, err := models.CreateObjectContent(tx, bucket.ID, key, util.RandomString(10), bytes)
			if err != nil {
				return err
			}

			version = object.Version
			etag = object.ETag
		} else {
			object, err := models.GetLatestObject(tx, bucket.ID, key)
			if err != nil && !gorm.IsRecordNotFoundError(err) {
				return err
			}
			if !gorm.IsRecordNotFoundError(err) {
				if err = tx.Delete(&object).Error; err != nil {
					return err
				}
			}

			object, err = models.CreateObjectContent(tx, bucket.ID, key, "null", bytes)
			if err != nil {
				return err
			}

			etag = object.ETag
		}
		return nil
	})

	return version, etag, err
}
