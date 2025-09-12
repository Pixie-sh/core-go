package database_models

import "github.com/pixie-sh/errors-go"

type DatabaseMigrationsIdentifier string

func (i DatabaseMigrationsIdentifier) String() string {
	return string(i)
}

var uniquesDBIdentifier = map[DatabaseMigrationsIdentifier]struct{}{}

func NewDatabaseMigrationsIdentifier(idName string) DatabaseMigrationsIdentifier {
	_, ok := uniquesDBIdentifier[DatabaseMigrationsIdentifier(idName)]
	if ok {
		panic(errors.New("DatabaseMigrationsIdentifier '%s' already exists", idName))
	}

	dbID := DatabaseMigrationsIdentifier(idName)
	uniquesDBIdentifier[dbID] = struct{}{}
	return dbID
}

type DatabaseMigratePayload struct {
	Identifiers   []DatabaseMigrationsIdentifier `json:"identifiers"`
	Transactional bool                           `json:"transactional"`
	Rollback      bool                           `json:"rollback,omitempty"`
	RollbackTo    string                         `json:"rollback_to,omitempty"` //rolls back up to migration id or $LAST
}

type DatabaseMigrateAllPayload struct {
	Transactional bool   `json:"transactional"`
	Rollback      bool   `json:"rollback,omitempty"`
	RollbackTo    string `json:"rollback_to,omitempty"` //rolls back up to migration id or $LAST
}

type BackofficeDatabaseExecutedMigrationsPayload struct {
	Identifiers []DatabaseMigrationsIdentifier `json:"identifiers"`
}
