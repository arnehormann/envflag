package value

import (
	"strconv"
	"time"
)

// Value is a value that can be retrieved or set in a string representation.
type Value interface {

	// Get retrieves a copy of the value
	Get() interface{}
	Set(string) error
	String() string
	AppendTo([]byte) []byte
}

// ValueOf retrieves a value referenced by ptr.
//
// The referenced value must be either be an int, uint or float type,
// or it must be a bool, string, or time.Duration.
//
// As in the flag package, IsBoolFlag() returns true for bool values.
func ValueOf(ptr interface{}) (val Value, ok bool) {
	if ptr == nil {
		return nil, false
	}
	switch val := ptr.(type) {
	case *string:
		return (*stringValue)(val), true
	case *bool:
		return (*boolValue)(val), true
	case *int:
		return (*intValue)(val), true
	case *int8:
		return (*int8Value)(val), true
	case *int16:
		return (*int16Value)(val), true
	case *int32:
		return (*int32Value)(val), true
	case *int64:
		return (*int64Value)(val), true
	case *uint:
		return (*uintValue)(val), true
	case *uint8:
		return (*uint8Value)(val), true
	case *uint16:
		return (*uint16Value)(val), true
	case *uint32:
		return (*uint32Value)(val), true
	case *uint64:
		return (*uint64Value)(val), true
	case *float32:
		return (*float32Value)(val), true
	case *float64:
		return (*float64Value)(val), true
	case *time.Duration:
		return (*durationValue)(val), true
	}
	return nil, false
}

type boolValue bool

func (p *boolValue) IsBoolFlag() bool { return true }
func (p *boolValue) Get() interface{} { return *(*bool)(p) }
func (p *boolValue) String() string   { return strconv.FormatBool(bool(*p)) }
func (p *boolValue) AppendTo(dest []byte) []byte {
	return strconv.AppendBool(dest, bool(*p))
}
func (p *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err == nil {
		*p = boolValue(v)
	}
	return err
}

type stringValue string

func (p *stringValue) Get() interface{} { return p.String() }
func (p *stringValue) String() string   { return *(*string)(p) }
func (p *stringValue) AppendTo(dest []byte) []byte {
	return append(dest, *(*string)(p)...)
}
func (p *stringValue) Set(s string) error {
	*p = stringValue(s)
	return nil
}

type uintValue uint

func (p *uintValue) Get() interface{} { return uint(*p) }
func (p *uintValue) String() string   { return strconv.FormatUint(uint64(*p), 10) }
func (p *uintValue) AppendTo(dest []byte) []byte {
	return strconv.AppendUint(dest, uint64(*p), 10)
}
func (p *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 0)
	if err == nil {
		*p = uintValue(v)
	}
	return err
}

type uint8Value uint8

func (p *uint8Value) Get() interface{} { return uint8(*p) }
func (p *uint8Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }
func (p *uint8Value) AppendTo(dest []byte) []byte {
	return strconv.AppendUint(dest, uint64(*p), 10)
}
func (p *uint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 8)
	if err == nil {
		*p = uint8Value(v)
	}
	return err
}

type uint16Value uint16

func (p *uint16Value) Get() interface{} { return uint16(*p) }
func (p *uint16Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }
func (p *uint16Value) AppendTo(dest []byte) []byte {
	return strconv.AppendUint(dest, uint64(*p), 10)
}
func (p *uint16Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 16)
	if err == nil {
		*p = uint16Value(v)
	}
	return err
}

type uint32Value uint32

func (p *uint32Value) Get() interface{} { return uint32(*p) }
func (p *uint32Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }
func (p *uint32Value) AppendTo(dest []byte) []byte {
	return strconv.AppendUint(dest, uint64(*p), 10)
}
func (p *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 32)
	if err == nil {
		*p = uint32Value(v)
	}
	return err
}

type uint64Value uint64

func (p *uint64Value) Get() interface{} { return uint64(*p) }
func (p *uint64Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }
func (p *uint64Value) AppendTo(dest []byte) []byte {
	return strconv.AppendUint(dest, uint64(*p), 10)
}
func (p *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*p = uint64Value(v)
	}
	return err
}

type intValue int

func (p *intValue) Get() interface{} { return int(*p) }
func (p *intValue) String() string   { return strconv.FormatInt(int64(*p), 10) }
func (p *intValue) AppendTo(dest []byte) []byte {
	return strconv.AppendInt(dest, int64(*p), 10)
}
func (p *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 0)
	if err == nil {
		*p = intValue(v)
	}
	return err
}

type int8Value int8

func (p *int8Value) Get() interface{} { return int8(*p) }
func (p *int8Value) String() string   { return strconv.FormatInt(int64(*p), 10) }
func (p *int8Value) AppendTo(dest []byte) []byte {
	return strconv.AppendInt(dest, int64(*p), 10)
}
func (p *int8Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 8)
	if err == nil {
		*p = int8Value(v)
	}
	return err
}

type int16Value int16

func (p *int16Value) Get() interface{} { return int16(*p) }
func (p *int16Value) String() string   { return strconv.FormatInt(int64(*p), 10) }
func (p *int16Value) AppendTo(dest []byte) []byte {
	return strconv.AppendInt(dest, int64(*p), 10)
}
func (p *int16Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 16)
	if err == nil {
		*p = int16Value(v)
	}
	return err
}

type int32Value int32

func (p *int32Value) Get() interface{} { return int32(*p) }
func (p *int32Value) String() string   { return strconv.FormatInt(int64(*p), 10) }
func (p *int32Value) AppendTo(dest []byte) []byte {
	return strconv.AppendInt(dest, int64(*p), 10)
}
func (p *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	if err == nil {
		*p = int32Value(v)
	}
	return err
}

type int64Value int64

func (p *int64Value) Get() interface{} { return int64(*p) }
func (p *int64Value) String() string   { return strconv.FormatInt(int64(*p), 10) }
func (p *int64Value) AppendTo(dest []byte) []byte {
	return strconv.AppendInt(dest, int64(*p), 10)
}
func (p *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*p = int64Value(v)
	}
	return err
}

type float32Value float32

func (p *float32Value) Get() interface{} { return float32(*p) }
func (p *float32Value) String() string   { return strconv.FormatFloat(float64(*p), 'g', -1, 32) }
func (p *float32Value) AppendTo(dest []byte) []byte {
	return strconv.AppendFloat(dest, float64(*p), 'g', -1, 32)
}
func (p *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	if err == nil {
		*p = float32Value(v)
	}
	return err
}

type float64Value float64

func (p *float64Value) Get() interface{} { return float64(*p) }
func (p *float64Value) String() string   { return strconv.FormatFloat(float64(*p), 'g', -1, 64) }
func (p *float64Value) AppendTo(dest []byte) []byte {
	return strconv.AppendFloat(dest, float64(*p), 'g', -1, 64)
}
func (p *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		*p = float64Value(v)
	}
	return err
}

type durationValue time.Duration

func (p *durationValue) Get() interface{} { return time.Duration(*p) }
func (p *durationValue) String() string   { return (*(*time.Duration)(p)).String() }
func (p *durationValue) AppendTo(dest []byte) []byte {
	return append(dest, (*(*time.Duration)(p)).String()...)
}
func (p *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err == nil {
		*p = durationValue(v)
	}
	return err
}
