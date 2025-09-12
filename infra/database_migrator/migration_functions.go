package bo_services

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/models/database_models"
)

func HandleMigrateWithService(ctx context.Context, migrator DatabaseMigratorService, command database_models.DatabaseMigratePayload) error {
	for _, identifier := range command.Identifiers {
		var err error
		switch command.Rollback {
		case false:
			err = migrator.Migrate(ctx, identifier, command.Transactional)
		case true:
			err = migrator.Rollback(ctx, identifier, command.Transactional, command.RollbackTo)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func HandleLooseMigrationWithService(
	ctx context.Context,
	migrator DatabaseMigratorService,
	command database_models.DatabaseMigrateAllPayload,
	migrations ...*database.Migration,
) error {
	var err error
	switch command.Rollback {
	case false:
		err = migrator.MigrateLoose(ctx, command.Transactional, migrations...)
	case true:
		panic("rollback loose not implemented yet")
		//err = migrator.RollbackLoose(ctx, command.Transactional, command.RollbackTo, migrations...) //TODO
	}

	if err != nil {
		return err
	}

	return nil
}

func MigratorOrderedListOf(registeredDatabaseMigrations map[database_models.DatabaseMigrationsIdentifier][]*database.Migration) []*database.Migration {
	var allMigrations []*database.Migration
	for _, migrations := range registeredDatabaseMigrations {
		allMigrations = append(allMigrations, migrations...)
	}

	sort.Slice(allMigrations, func(i, j int) bool {
		idPartsI := strings.SplitN(allMigrations[i].ID, "_", 2)
		idPartsJ := strings.SplitN(allMigrations[j].ID, "_", 2)

		epochI, errI := strconv.ParseInt(idPartsI[0], 10, 64)
		epochJ, errJ := strconv.ParseInt(idPartsJ[0], 10, 64)
		if errI != nil || errJ != nil {
			panic(errors.Join(errors.New("Unable to parse migration ID"), errI, errJ))
		}

		// Compare epochs
		return epochI < epochJ
	})

	return allMigrations
}
