package controllers

import (
	"net/http"
	_ "net/http/pprof"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/s2"
	"github.com/pachyderm/s2/examples/sql/models"
)

func (c *Controller) GetLocation(r *http.Request, name string) (string, error) {
	c.logger.Tracef("GetLocation: %+v", name)
	return models.Location, nil
}

// Lists bucket contents. Note that this doesn't support common prefixes or
// delimiters.
func (c *Controller) ListObjects(r *http.Request, name, prefix, marker, delimiter string, maxKeys int) (*s2.ListObjectsResult, error) {
	c.logger.Tracef("ListObjects: name=%+v, prefix=%+v, marker=%+v, delimiter=%+v, maxKeys=%+v", name, prefix, marker, delimiter, maxKeys)

	result := s2.ListObjectsResult{
		Contents:       []*s2.Contents{},
		CommonPrefixes: []*s2.CommonPrefixes{},
	}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		objects, err := models.ListObjects(tx, bucket.ID, marker, "", maxKeys+1)
		if err != nil {
			return err
		}

		// mapping of already listed objects, so we don't include multiple
		// versions of the same object
		covered := map[string]bool{}

		for _, object := range objects {
			if !strings.HasPrefix(object.Key, prefix) || isDelimiterFiltered(object.Key, prefix, delimiter) {
				continue
			}

			if delimiter != "" {
				parts := strings.SplitN(object.Key[len(prefix):], delimiter, 2)
				if len(parts) == 2 && len(parts[1]) > 0 {
					continue
				}
			}

			if _, ok := covered[object.Key]; ok {
				continue
			}

			latestObject, err := models.GetLatestObject(tx, bucket.ID, object.Key)
			if err != nil {
				return err
			}
			covered[latestObject.Key] = true
			if latestObject.DeleteMarker {
				continue
			}

			if len(result.Contents)+len(result.CommonPrefixes) >= maxKeys {
				if maxKeys > 0 {
					result.IsTruncated = true
				}
				break
			}

			result.Contents = append(result.Contents, &s2.Contents{
				Key:          latestObject.Key,
				LastModified: models.Epoch,
				ETag:         latestObject.ETag,
				Size:         uint64(len(latestObject.Content)),
				StorageClass: models.StorageClass,
				Owner:        models.GlobalUser,
			})
		}

		return nil
	})

	return &result, err
}

func (c *Controller) ListObjectVersions(r *http.Request, name, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (*s2.ListObjectVersionsResult, error) {
	c.logger.Tracef("ListObjectVersions: name=%+v, prefix=%+v, keyMarker=%+v, versionMarker=%+v, delimiter=%+v, maxKeys=%+v", name, prefix, keyMarker, versionMarker, delimiter, maxKeys)

	result := s2.ListObjectVersionsResult{
		Versions:      []*s2.Version{},
		DeleteMarkers: []*s2.DeleteMarker{},
	}

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		objects, err := models.ListObjects(tx, bucket.ID, keyMarker, versionMarker, maxKeys+1)
		if err != nil {
			return err
		}

		// s3tests expects the listings to be ordered by update date
		sort.Slice(objects, func(i, j int) bool {
			return objects[i].ID > objects[j].ID
		})

		for _, object := range objects {
			if !strings.HasPrefix(object.Key, prefix) || isDelimiterFiltered(object.Key, prefix, delimiter) {
				continue
			}

			if len(result.Versions)+len(result.DeleteMarkers) >= maxKeys {
				if maxKeys > 0 {
					result.IsTruncated = true
				}
				break
			}

			latestObject, err := models.GetLatestObject(tx, bucket.ID, object.Key)
			if err != nil {
				return err
			}

			if object.DeleteMarker {
				result.DeleteMarkers = append(result.DeleteMarkers, &s2.DeleteMarker{
					Key:          object.Key,
					Version:      object.Version,
					IsLatest:     latestObject.ID == object.ID,
					LastModified: models.Epoch,
					Owner:        models.GlobalUser,
				})
			} else {
				result.Versions = append(result.Versions, &s2.Version{
					Key:          object.Key,
					Version:      object.Version,
					IsLatest:     latestObject.ID == object.ID,
					LastModified: models.Epoch,
					ETag:         object.ETag,
					Size:         uint64(len(object.Content)),
					StorageClass: models.StorageClass,
					Owner:        models.GlobalUser,
				})
			}
		}

		return nil
	})

	return &result, err
}

func (c *Controller) CreateBucket(r *http.Request, name string) error {
	c.logger.Tracef("CreateBucket: %+v", name)

	return c.transaction(func(tx *gorm.DB) error {
		_, err := models.GetBucket(tx, name)
		if err == nil {
			return s2.BucketAlreadyOwnedByYouError(r)
		} else if !gorm.IsRecordNotFoundError(err) {
			return err
		}

		_, err = models.CreateBucket(tx, name)
		return err
	})
}

func (c *Controller) DeleteBucket(r *http.Request, name string) error {
	c.logger.Tracef("DeleteBucket: %+v", name)

	return c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		return tx.Delete(bucket).Error
	})
}

func (c *Controller) GetBucketVersioning(r *http.Request, name string) (string, error) {
	c.logger.Tracef("GetBucketVersioning: %+v", name)

	result := ""

	err := c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		result = bucket.Versioning
		return nil
	})

	return result, err
}

func (c *Controller) SetBucketVersioning(r *http.Request, name, status string) error {
	c.logger.Tracef("SetBucketVersioning: name=%+v, status=%+v", name, status)

	return c.transaction(func(tx *gorm.DB) error {
		bucket, err := models.GetBucket(tx, name)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return s2.NoSuchBucketError(r)
			}
			return err
		}

		if status == s2.VersioningDisabled && bucket.Versioning != "" {
			return s2.IllegalVersioningConfigurationError(r)
		}

		bucket.Versioning = status

		return tx.Save(&bucket).Error
	})
}

func isDelimiterFiltered(key, prefix, delimiter string) bool {
	if delimiter == "" {
		return false
	}
	parts := strings.SplitN(key[len(prefix):], delimiter, 2)
	return len(parts) == 2 && len(parts[1]) > 0
}
