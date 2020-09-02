package gc

import (
	"time"

	"github.com/jinzhu/gorm"
)

var (
	refTable   = "refs"
	chunkTable = "chunks"
)

type refModel struct {
	Sourcetype string `sql:"type:reftype" gorm:"primary_key"`
	Source     string `gorm:"primary_key"`
	Chunk      string `gorm:"primary_key"`
	Created    *time.Time
	ExpiresAt  *time.Time
}

func (*refModel) TableName() string {
	return refTable
}

type chunkModel struct {
	Chunk    string `gorm:"primary_key"`
	Deleting *time.Time
}

func (*chunkModel) TableName() string {
	return chunkTable
}

func convertChunks(chunks []chunkModel) []string {
	result := make([]string, 0, len(chunks))
	for _, c := range chunks {
		result = append(result, c.Chunk)
	}
	return result
}

func initializeDb(db *gorm.DB) error {
	// Make sure the `reftype` enum exists
	if err := db.Exec(`
DO $$ 
BEGIN
	CREATE TYPE reftype AS ENUM 
	('chunk', 'temporary', 'semantic');
EXCEPTION
	WHEN duplicate_object THEN NULL;
END $$
  `).Error; err != nil {
		return err
	}

	if err := db.AutoMigrate(&refModel{}, &chunkModel{}).Error; err != nil {
		return err
	}

	if err := db.Model(&refModel{}).AddIndex("idx_chunk", "chunk").Error; err != nil {
		return err
	}

	if err := db.Model(&refModel{}).AddIndex("idx_sourcetype_source", "sourcetype", "source").Error; err != nil {
		return err
	}

	return nil
}
