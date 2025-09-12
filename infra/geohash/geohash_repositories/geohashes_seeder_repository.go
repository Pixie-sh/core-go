package geohash_repositories

import (
	"gorm.io/gorm/clause"

	"github.com/pixie-sh/database-helpers-go/database"
)

type GeohashSeederRepository struct {
	database.Repository[GeohashSeederRepository]
}

func NewGeohashSeederRepository(db *database.DB) GeohashSeederRepository {
	return GeohashSeederRepository{database.NewRepository(db, NewGeohashSeederRepository)}
}

func (r GeohashSeederRepository) GetByHash(geohash string) (sm Geohash, e error) {
	return sm, r.DB.Model(&Geohash{}).
		Where("geohash", geohash).
		First(&sm).
		Error
}

func (r GeohashSeederRepository) BatchInsert(batch []Geohash) ([]Geohash, error) {
	return batch, r.DB.Model(&Geohash{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "geohash"}},
			DoUpdates: clause.AssignmentColumns([]string{"iso3"}),
		}).
		CreateInBatches(batch, len(batch)).Error
}

func (r GeohashSeederRepository) TruncateTable() error {
	return r.DB.Exec("TRUNCATE TABLE geohashes RESTART IDENTITY CASCADE").Error
}
