package controllers

import (
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	logger *logrus.Entry
	db     *gorm.DB
	lock   *sync.Mutex
}

func NewController(logger *logrus.Entry, db *gorm.DB) *Controller {
	return &Controller{
		logger: logger,
		db:     db,
		lock:   &sync.Mutex{},
	}
}

// transaction calls a given function with a database transaction. If the
// function returns an error, the transaction is rolled back. Otherwise
// it's committed.
func (c *Controller) transaction(f func(*gorm.DB) error) error {
	// This is unfortunate. gorm doesn't offer access sqlite3's locking
	// mechanism, and sqlite3's default is to either serve table missing
	// errors, or lock errors, when multiple transactions try to write to the
	// database at the same time. So we have to wrap everything in a global
	// lock.
	c.lock.Lock()
	defer c.lock.Unlock()

	tx := c.db.New().Begin()
	err := f(tx)
	if err != nil {
		if err := tx.Rollback().Error; err != nil {
			c.logger.WithError(err).Error("could not rollback transaction")
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		c.logger.WithError(err).Error("could not commit transaction")
	}
	return nil
}
