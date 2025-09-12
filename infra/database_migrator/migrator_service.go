package bo_services

import (
	"context"

	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/models/database_models"
	"github.com/pixie-sh/core-go/pkg/types/maps"
)

type DatabaseMigratorService struct {
	registeredDatabaseMigrations map[database_models.DatabaseMigrationsIdentifier][]*database.Migration
	db                           *database.DB
}

func NewDatabaseMigratorService(
	_ context.Context,
	db *database.DB,
	registeredDatabaseMigrations map[database_models.DatabaseMigrationsIdentifier][]*database.Migration,
) DatabaseMigratorService {
	dms := DatabaseMigratorService{
		db:                           db,
		registeredDatabaseMigrations: registeredDatabaseMigrations,
	}

	return dms
}

func (m DatabaseMigratorService) Migrate(
	ctx context.Context,
	serviceIdentifier database_models.DatabaseMigrationsIdentifier,
	transactional bool,
) error {
	migrations := m.getServiceMigrations(serviceIdentifier)
	if migrations == nil {
		return errors.New("no migrations available for service %s", serviceIdentifier).WithErrorCode(errors.NotFoundErrorCode)
	}

	return m.MigrateLoose(ctx, transactional, migrations...)
}

func (m DatabaseMigratorService) Rollback(ctx context.Context, serviceIdentifier database_models.DatabaseMigrationsIdentifier, transactional bool, rollbackToMigrationID string) error {
	if len(rollbackToMigrationID) == 0 {
		return errors.New("cannot rollback with empty rollback_to. to rollback last use keyword: $LAST")
	}

	var migratorConfig = database.MigratorConfiguration{
		UseTransaction: transactional,
	}

	migrations := m.getServiceMigrations(serviceIdentifier)
	if migrations == nil {
		return errors.New("no migrations available for service %s", serviceIdentifier).WithErrorCode(errors.NotFoundErrorCode)
	}

	migrator, err := database.NewMigrator(ctx, &migratorConfig, m.db, migrations...)
	if err != nil {
		return err
	}

	if rollbackToMigrationID == "$LAST" {
		return migrator.RollbackLast()
	}

	return migrator.RollbackTo(rollbackToMigrationID)
}

func (m DatabaseMigratorService) getServiceMigrations(identifier database_models.DatabaseMigrationsIdentifier) []*database.Migration {
	migrations, ok := m.registeredDatabaseMigrations[identifier]
	if !ok {
		return nil
	}

	return migrations
}

func (m DatabaseMigratorService) ListExecuted(_ context.Context) ([]string, error) {
	type migration struct {
		ID string `json:"id"`
	}

	var migrations []migration

	result := m.db.Table("migrations").Find(&migrations)
	if result.Error != nil {
		return nil, result.Error
	}

	return maps.MapStructValue[migration, string](migrations, func(m migration) string {
		return m.ID
	}), nil
}

func (m DatabaseMigratorService) MigrateLoose(ctx context.Context, transactional bool, migrations ...*database.Migration) error {
	var migratorConfig = database.MigratorConfiguration{
		UseTransaction: transactional,
	}

	if !transactional {
		for _, migration := range migrations {
			pixiecontext.GetCtxLogger(ctx).With("migration", migration).Debug("Starting Migration %s", migration.ID)

			migrator, err := database.NewMigrator(ctx, &migratorConfig, m.db, migrations...)
			if err != nil {
				return err
			}

			err = migrator.Migrate()
			if err != nil {
				return err
			}
		}
	} else {
		pixiecontext.GetCtxLogger(ctx).With("migration", migrations).Debug("Starting Migration Bulk len(%d)", len(migrations))
		migrator, err := database.NewMigrator(ctx, &migratorConfig, m.db, migrations...)
		if err != nil {
			return err
		}

		err = migrator.Migrate()
		if err != nil {
			return err
		}
	}

	return nil
}
