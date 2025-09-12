package geohash

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/infra/geohash/geohash_repositories"
	"github.com/pixie-sh/core-go/pkg/services"
)

//go:embed master-data.csv.gz
var masterDataCSV []byte

type GeohashSeederServiceConfiguration struct {
}

type GeohashSeederService struct {
	services.Service[GeohashSeederService]

	config GeohashSeederServiceConfiguration
	repo   geohash_repositories.GeohashSeederRepository
}

func NewGeohashSeederService(_ context.Context, config GeohashSeederServiceConfiguration, repo geohash_repositories.GeohashSeederRepository) (GeohashSeederService, error) {
	is := GeohashSeederService{
		config: config,
		repo:   repo,
	}

	is.Service = services.NewService(&is, NewGeohashSeederServiceFrom)
	return is, nil
}

func NewGeohashSeederServiceFrom(service services.Service[GeohashSeederService]) (*GeohashSeederService, *services.Service[GeohashSeederService]) {
	us := &GeohashSeederService{}
	us.config = service.Instance.config
	us.repo = service.Instance.repo
	us.Service = services.NewService(us, nil)

	return us, &us.Service
}

func (gs GeohashSeederService) SeedGeohashes(_ context.Context, truncate bool) error {
	if truncate {
		if err := gs.repo.TruncateTable(); err != nil {
			return errors.NewWithError(err, "failed to truncate geohash table: %w", err)
		}
	}

	reader, err := gzip.NewReader(bytes.NewReader(masterDataCSV))
	if err != nil {
		return errors.NewWithError(err, "failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	csvReader := csv.NewReader(reader)
	var batch []geohash_repositories.Geohash

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.NewWithError(err, "failed to read CSV record: %w", err)
		}

		var geohashEntity geohash_repositories.Geohash
		geohashEntity.Geohash = record[0]
		geohashEntity.ISO3 = record[1]
		batch = append(batch, geohashEntity)

		if len(batch) == 1000 {
			if _, err := gs.repo.BatchInsert(batch); err != nil {
				return errors.NewWithError(err, "failed to perform batch insert: %w", err)
			}

			batch = nil
		}
	}

	if len(batch) > 0 {
		if _, err := gs.repo.BatchInsert(batch); err != nil {
			return fmt.Errorf("failed to perform final batch insert: %w", err)
		}
	}

	return nil
}
