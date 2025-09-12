package geohash_repositories

import (
	"github.com/pixie-sh/database-helpers-go/database"
	"gorm.io/gorm"
)

type geohash1744897031888 struct {
	Geohash string `gorm:"type:text;primaryKey"`
	ISO3    string `gorm:"type:varchar(3);index"`
} //@name Geohash

func (geohash1744897031888) TableName() string {
	return "geohashes"
}

var CreateGeohashTable1744897031888 = database.Migration{
	ID: "1744897031888_CreateGeohashTable",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&geohash1744897031888{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
            DROP TABLE IF EXISTS geohashes;
        `).Error
	},
}
