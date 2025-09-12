package uid

import (
	pulid "github.com/pixie-sh/ulid-go"
)

type UID = pulid.ULID
type Scope = pulid.Scope

type UIDs []UID

func (uids UIDs) UUIDArray() []string {
	var uuids []string
	for _, uid := range uids {
		uuids = append(uuids, uid.UUID())
	}

	return uuids
}

var LowerBoundForbiddenScope = pulid.ZeroedScopeValue
var UpperBoundForbiddenScope = pulid.MaxScopeValue
var Nil = pulid.EmptyUID

// New return a UID compatible struct, time based with monotonic entropy
func New() UID {
	return pulid.MustNew()
}

// NewULID return ULID compatible string
func NewULID() string {
	return pulid.MustNew().String()
}

// NewUUID return UUID compatible string
func NewUUID() string {
	return pulid.MustNew().UUID()
}

// FromString return UID from uuid v4 or UID string
func FromString(s string) (UID, error) {
	id, err := pulid.UnmarshalString(s)
	if err != nil {
		return Nil, err
	}

	return id, nil
}

func FromUint64(u uint64) (UID, error) {
	id, err := pulid.UnmarshalUint64(u)
	if err != nil {
		return Nil, err
	}

	return id, nil
}

// MustString return UID from uuid v4 or UID string
// panics if error
func MustString(s string) UID {
	id, err := FromString(s)
	if err != nil {
		panic(err)
	}

	return id
}

// NewScoped generates a scoped UID with default entropy
func NewScoped(scope Scope) UID {
	return pulid.MustNewScoped(scope)
}
