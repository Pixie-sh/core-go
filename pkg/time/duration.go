package time

import (
	gotime "time"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

type Duration gotime.Duration

// ParseDuration mimic go time.ParseDuration
func ParseDuration[vv float64 | string](v vv) (Duration, error) {
	return NewDuration(v)
}

// NewDuration creates a Duration from either a string or float64 value
func NewDuration[vv float64 | string](v vv) (Duration, error) {
	var vIf interface{} = v
	switch value := vIf.(type) {
	case float64:
		return Duration(gotime.Duration(value)), nil
	case string:
		duration, err := gotime.ParseDuration(value)
		if err != nil {
			return Duration(0), err
		}
		return Duration(duration), nil
	default:
		return Duration(0), errors.New("invalid duration type")
	}
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return serializer.Serialize(gotime.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := serializer.Deserialize(b, &v, false)

	if err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*d = Duration(gotime.Duration(value))
		return nil
	case string:
		duration, err := gotime.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(duration)
		return nil
	default:
		return errors.New("invalid duration type")
	}
}

// Duration Helper method to get the time.Duration
func (d Duration) Duration() gotime.Duration {
	return gotime.Duration(d)
}

// String Helper method to get the string representation
func (d Duration) String() string {
	return gotime.Duration(d).String()
}

// FromDuration creates Duration from time.Duration
func (d *Duration) FromDuration(goDuration gotime.Duration) Duration {
	*d = Duration(goDuration)
	return *d
}

// FromFloat64 creates Duration from float64
func (d *Duration) FromFloat64(f float64) Duration {
	*d = Duration(gotime.Duration(f))
	return *d
}

func (d Duration) Hours() float64 {
	return d.Duration().Hours()
}

func (d Duration) Minutes() float64 {
	return d.Duration().Minutes()
}

func (d Duration) Seconds() float64 {
	return d.Duration().Seconds()
}
