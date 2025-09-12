package geohash_repositories

import (
	"github.com/pixie-sh/database-helpers-go/database"

	"github.com/pixie-sh/core-go/pkg/geometry"
)

type GeohashesRepository struct {
	database.Repository[GeohashesRepository]
}

func NewGeohashesRepository(db *database.DB) GeohashesRepository {
	return GeohashesRepository{database.NewRepository(db, NewGeohashesRepository)}
}

func (r GeohashesRepository) GetByGeohash(geohash string) (sm Geohash, e error) {
	return sm, r.DB.Model(&Geohash{}).
		Where("geohash", geohash).
		First(&sm).
		Error
}

func (r GeohashesRepository) GetByGeohashes(geohashes []geometry.Geohash) (sm []Geohash, e error) {
	return sm, r.DB.Model(&Geohash{}).
		Where("geohash IN (?)", geohashes).
		Find(&sm).
		Error
}
