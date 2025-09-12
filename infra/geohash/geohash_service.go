package geohash

import (
	"context"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/infra/geohash/geohash_repositories"
	"github.com/pixie-sh/core-go/pkg/errors/db_errors"
	"github.com/pixie-sh/core-go/pkg/geometry"
	"github.com/pixie-sh/core-go/pkg/services"
)

type GeohashServiceConfiguration struct {
}

type GeohashService struct {
	services.Service[GeohashService]

	config GeohashServiceConfiguration
	repo   geohash_repositories.GeohashesRepository
}

func NewGeohashService(_ context.Context, config GeohashServiceConfiguration, repo geohash_repositories.GeohashesRepository) (GeohashService, error) {
	is := GeohashService{
		config: config,
		repo:   repo,
	}

	is.Service = services.NewService(&is, NewGeohashServiceFrom)
	return is, nil
}

func NewGeohashServiceFrom(service services.Service[GeohashService]) (*GeohashService, *services.Service[GeohashService]) {
	us := &GeohashService{}
	us.config = service.Instance.config
	us.repo = service.Instance.repo
	us.Service = services.NewService(us, nil)

	return us, &us.Service
}

func (l GeohashService) FindGeohashesForCoordinates(_ context.Context, lat float64, lng float64) ([]geohash_repositories.Geohash, error) {
	latLongGeoHash := geometry.NewGeohash().FromLatLon(lat, lng, 6)
	txRepo := l.repo
	if !l.TxNil() {
		txRepo = txRepo.WithTx(l.Tx)
	}

	geohashes, err := txRepo.GetByGeohashes(latLongGeoHash.PrecisionVariations())
	if err != nil {
		return nil, db_errors.Handle(err)
	}

	if len(geohashes) == 0 {
		return nil, errors.NewValidationError("iso3 not found", &errors.FieldError{
			Rule:    "entityNotFound",
			Message: "No ISO3 found for the given coordinates",
		}).WithErrorCode(errors.EntityNotFoundErrorCode)
	}

	biggestGeohash := geohashes[0]
	for _, geohash := range geohashes {
		if len(biggestGeohash.Geohash) < len(geohash.Geohash) {
			biggestGeohash = geohash
		}
	}

	return geohashes, nil
}

func (l GeohashService) FindGeohashForCoordinates(_ context.Context, lat float64, lng float64) (geohash_repositories.Geohash, error) {
	txRepo := l.repo
	if !l.TxNil() {
		txRepo = txRepo.WithTx(l.Tx)
	}

	latLongGeoHash := geometry.NewGeohash().FromLatLon(lat, lng, 6)
	geohashes, err := txRepo.GetByGeohashes(latLongGeoHash.PrecisionVariations())
	if err != nil {
		return geohash_repositories.Geohash{}, db_errors.Handle(err)
	}

	if len(geohashes) == 0 {
		return geohash_repositories.Geohash{}, errors.NewValidationError("iso3 not found", &errors.FieldError{
			Rule:    "entityNotFound",
			Message: "No ISO3 found for the given coordinates",
		}).WithErrorCode(errors.EntityNotFoundErrorCode)
	}

	biggestGeohash := geohashes[0]
	for _, geohash := range geohashes {
		if len(biggestGeohash.Geohash) < len(geohash.Geohash) {
			biggestGeohash = geohash
		}
	}

	return biggestGeohash, nil
}

func (l GeohashService) FindISO3ForCoordinates(_ context.Context, lat float64, lng float64) (string, error) {
	latLngGeohash := geometry.NewGeohash().FromLatLon(lat, lng, 6)

	txRepo := l.repo
	if !l.TxNil() {
		txRepo = txRepo.WithTx(l.Tx)
	}

	geohashes, err := txRepo.GetByGeohashes(latLngGeohash.PrecisionVariations())
	if err != nil {
		return "", db_errors.Handle(err)
	}

	if len(geohashes) == 0 {
		return "", errors.NewValidationError("iso3 not found", &errors.FieldError{
			Rule:    "entityNotFound",
			Message: "No ISO3 found for the given coordinates",
		}).WithErrorCode(errors.EntityNotFoundErrorCode)
	}

	biggestGeohash := geohashes[0]
	for _, geohash := range geohashes {
		if len(biggestGeohash.Geohash) < len(geohash.Geohash) {
			biggestGeohash = geohash
		}
	}

	return biggestGeohash.ISO3, nil
}
