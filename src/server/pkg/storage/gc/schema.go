package gc

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

var (
	refTable   = "refs"
	chunkTable = "chunks"
)

type refModel struct {
	Sourcetype string `sql:"type:reftype" gorm:"primary_key"`
	Source     string `gorm:"primary_key"`
	Chunk      string `gorm:"primary_key"`
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

func convertChunks(chunks []chunkModel) []chunk.Chunk {
	result := make([]chunk.Chunk, 0, len(chunks))
	for _, c := range chunks {
		result = append(result, chunk.Chunk{Hash: c.Chunk})
	}
	return result
}

func initializeDb(db *gorm.DB) error {
	// Make sure the `reftype` enum exists
	if err := db.Exec(`
do $$ begin
 create type reftype as enum ('chunk', 'job', 'semantic');
exception
 when duplicate_object then null;
end $$
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
