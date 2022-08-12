package controllers

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/s2"
	"github.com/pachyderm/s2/examples/sql/models"
)

func (c *Controller) ListBuckets(r *http.Request) (*s2.ListBucketsResult, error) {
	c.logger.Tracef("ListBuckets")

	result := s2.ListBucketsResult{
		Owner:   &models.GlobalUser,
		Buckets: []*s2.Bucket{},
	}

	err := c.transaction(func(tx *gorm.DB) error {
		var buckets []*models.Bucket
		if err := tx.Find(&buckets).Error; err != nil {
			return err
		}

		for _, bucket := range buckets {
			result.Buckets = append(result.Buckets, &s2.Bucket{
				Name:         bucket.Name,
				CreationDate: models.Epoch,
			})
		}

		return nil
	})

	return &result, err
}
