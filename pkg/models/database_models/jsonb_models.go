package database_models

import (
	"database/sql/driver"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

// JSONB just wrapper for jsonb format
// GORM doc is using Scan with ptr and Value without ptr
// https://gorm.io/docs/data_types.html
type JSONB map[string]interface{}

// SliceJSONB just wrapper for list of jsonb format
// GORM doc is using Scan with ptr and Value without ptr
// https://gorm.io/docs/data_types.html
type SliceJSONB []JSONB

func (j *SliceJSONB) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("PayloadType assertion .([]byte) failed.")
	}

	var i interface{}
	if err := serializer.Deserialize(source, &i, false); err != nil {
		return err
	}

	ii, ok := i.([]interface{})
	if !ok {
		return errors.New("PayloadType assertion .([]interface{}) failed.")
	}

	*j = make([]JSONB, len(ii), len(ii))
	for idx := range ii {
		jb := JSONB{}
		err := jb.scanInterface(ii[idx])
		if err != nil {
			return errors.NewWithError(err, "Unable to map JSONB.")
		}

		(*j)[idx] = jb
	}

	return nil
}

func (j SliceJSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}

	return serializer.Serialize(j)
}

func (j SliceJSONB) GormDataType() string {
	return "jsonb"
}

func (j *JSONB) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("PayloadType assertion .([]byte) failed.")
	}

	var i interface{}
	if err := serializer.Deserialize(source, &i, false); err != nil {
		return err
	}

	*j, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("PayloadType assertion .(map[string]interface{}) failed.")
	}

	return nil
}

func (j *JSONB) scanInterface(source interface{}) error {
	var ok = false
	*j, ok = source.(map[string]interface{})
	if !ok {
		return errors.New("PayloadType assertion .(map[string]interface{}) failed.")
	}

	return nil
}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return serializer.Serialize(j)
}

func (j JSONB) GormDataType() string {
	return "jsonb"
}
