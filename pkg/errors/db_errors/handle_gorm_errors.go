package db_errors

import (
	goErrors "errors"
	"strings"

	"github.com/pixie-sh/errors-go"
	"gorm.io/gorm"
)

var DBEntityNotFound = gorm.ErrRecordNotFound
var DBInvalidTransaction = gorm.ErrInvalidTransaction
var DBNotImplemented = gorm.ErrNotImplemented
var DBMissingWhereClause = gorm.ErrMissingWhereClause
var DBUnsupportedRelation = gorm.ErrUnsupportedRelation
var DBPrimaryKeyRequired = gorm.ErrPrimaryKeyRequired
var DBModelValueRequired = gorm.ErrModelValueRequired
var DBModelAccessibleFieldsRequired = gorm.ErrModelAccessibleFieldsRequired
var DBSubQueryRequired = gorm.ErrSubQueryRequired
var DBInvalidData = gorm.ErrInvalidData
var DBUnsupportedDriver = gorm.ErrUnsupportedDriver
var DBRegistered = gorm.ErrRegistered
var DBInvalidField = gorm.ErrInvalidField
var DBEmptySlice = gorm.ErrEmptySlice
var DBDryRunModeUnsupported = gorm.ErrDryRunModeUnsupported
var DBInvalidDB = gorm.ErrInvalidDB
var DBInvalidValue = gorm.ErrInvalidValue
var DBInvalidValueOfLength = gorm.ErrInvalidValueOfLength
var DBPreloadNotAllowed = gorm.ErrPreloadNotAllowed
var DBDuplicatedKey = gorm.ErrDuplicatedKey
var DBForeignKeyViolated = gorm.ErrForeignKeyViolated
var DBCheckConstraintViolated = gorm.ErrCheckConstraintViolated

// handleGormError processes a given GORM error and maps it to a custom error type.
// If the error is not recognized or not explicitly handled, it returns nil.
func handleGormError(e error) (error, bool) {
	if goErrors.Is(e, DBEntityNotFound) {
		return errors.NewValidationError("record not found", &errors.FieldError{
			Rule:    "entityNotFound",
			Message: e.Error(),
		}).WithErrorCode(errors.EntityNotFoundErrorCode), true
	}

	if goErrors.Is(e, DBInvalidTransaction) {
		return errors.New("invalid transaction", &errors.FieldError{
			Rule:    "invalidTransaction",
			Message: e.Error(),
		}, errors.DBInvalidTransactionErrorCode), true
	}

	if goErrors.Is(e, DBNotImplemented) {
		return errors.New("not implemented", &errors.FieldError{
			Rule:    "notImplemented",
			Message: e.Error(),
		}, errors.DBNotImplementedErrorCode), true
	}

	if goErrors.Is(e, DBMissingWhereClause) {
		return errors.New("missing WHERE clause", &errors.FieldError{
			Rule:    "missingWhereClause",
			Message: e.Error(),
		}, errors.QueryMissingWhereClauseErrorCode), true
	}

	if goErrors.Is(e, DBUnsupportedRelation) {
		return errors.New("unsupported relation", &errors.FieldError{
			Rule:    "unsupportedRelation",
			Message: e.Error(),
		}, errors.QueryUnsupportedRelationErrorCode), true
	}

	if goErrors.Is(e, DBPrimaryKeyRequired) {
		return errors.New("primary key required", &errors.FieldError{
			Rule:    "primaryKeyRequired",
			Message: e.Error(),
		}, errors.QueryPrimaryKeyRequiredErrorCode), true
	}

	if goErrors.Is(e, DBModelValueRequired) {
		return errors.New("model value required", &errors.FieldError{
			Rule:    "modelValueRequired",
			Message: e.Error(),
		}, errors.EntityModelValueRequiredErrorCode), true
	}

	if goErrors.Is(e, DBModelAccessibleFieldsRequired) {
		return errors.New("model accessible fields required", &errors.FieldError{
			Rule:    "modelAccessibleFieldsRequired",
			Message: e.Error(),
		}, errors.EntityModelAccessibleFieldsRequiredErrorCode), true
	}

	if goErrors.Is(e, DBSubQueryRequired) {
		return errors.New("subquery required", &errors.FieldError{
			Rule:    "subQueryRequired",
			Message: e.Error(),
		}, errors.QuerySubQueryRequiredErrorCode), true
	}

	if goErrors.Is(e, DBInvalidData) {
		return errors.New("invalid data", &errors.FieldError{
			Rule:    "invalidData",
			Message: e.Error(),
		}, errors.QueryInvalidDataErrorCode), true
	}

	if goErrors.Is(e, DBUnsupportedDriver) {
		return errors.New("unsupported driver", &errors.FieldError{
			Rule:    "unsupportedDriver",
			Message: e.Error(),
		}, errors.DBUnsupportedDriverErrorCode), true
	}

	if goErrors.Is(e, DBRegistered) {
		return errors.New("already registered", &errors.FieldError{
			Rule:    "alreadyRegistered",
			Message: e.Error(),
		}, errors.DBRegisteredErrorCode), true
	}

	if goErrors.Is(e, DBInvalidField) {
		return errors.New("invalid field", &errors.FieldError{
			Rule:    "invalidField",
			Message: e.Error(),
		}, errors.QueryInvalidFieldErrorCode), true
	}

	if goErrors.Is(e, DBEmptySlice) {
		return errors.New("empty slice found", &errors.FieldError{
			Rule:    "emptySlice",
			Message: e.Error(),
		}, errors.EntityEmptySliceErrorCode), true
	}

	if goErrors.Is(e, DBDryRunModeUnsupported) {
		return errors.New("dry run mode not supported", &errors.FieldError{
			Rule:    "dryRunModeUnsupported",
			Message: e.Error(),
		}, errors.DBDryRunModeUnsupportedErrorCode), true
	}

	if goErrors.Is(e, DBInvalidDB) {
		return errors.New("invalid database", &errors.FieldError{
			Rule:    "invalidDB",
			Message: e.Error(),
		}, errors.DBInvalidDatabaseErrorCode), true
	}

	if goErrors.Is(e, DBInvalidValue) {
		return errors.New("invalid value", &errors.FieldError{
			Rule:    "invalidValue",
			Message: e.Error(),
		}, errors.DBInvalidValueErrorCode), true
	}

	if goErrors.Is(e, DBInvalidValueOfLength) {
		return errors.New("invalid value length", &errors.FieldError{
			Rule:    "invalidValueLength",
			Message: e.Error(),
		}, errors.DBInvalidValueOfLengthErrorCode), true
	}

	if goErrors.Is(e, DBPreloadNotAllowed) {
		return errors.New("preload not allowed", &errors.FieldError{
			Rule:    "preloadNotAllowed",
			Message: e.Error(),
		}, errors.QueryPreloadNotAllowedErrorCode), true
	}

	if goErrors.Is(e, DBDuplicatedKey) {
		return errors.New("duplicated key", &errors.FieldError{
			Rule:    "duplicatedKey",
			Message: e.Error(),
		}, errors.QueryDuplicatedKeyErrorCode), true
	}

	if goErrors.Is(e, DBForeignKeyViolated) || strings.Contains(e.Error(), "violates foreign key constraint") {
		return errors.New("foreign key constraint violated", &errors.FieldError{
			Rule:    "foreignKeyViolated",
			Message: e.Error(),
		}, errors.EntityForeignKeyViolatedErrorCode), true
	}

	if goErrors.Is(e, DBCheckConstraintViolated) {
		return errors.New("check constraint violated", &errors.FieldError{
			Rule:    "checkConstraintViolated",
			Message: e.Error(),
		}, errors.QueryCheckConstraintViolatedErrorCode), true
	}

	return nil, false
}
