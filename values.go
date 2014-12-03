package envflag

import (
	"errors"
	"reflect"
	"strconv"
	"time"
)

// Value is a parameter that can be converted to and from string.
//
// It matches flag.Value.
type Value interface {
	String() string
	Set(string) error
}

// ValueOf provides a Value for a pointer.
//
// Unless the pointer ptr or its destination implement Value,
// ptr must point to an int or uint type,
// bool, byte, float32, float64, string or time.Duration.
func ValueOf(ptr interface{}) (Value, error) {
	if ptr == nil {
		return nil, errors.New("ptr is nil")
	}
	var value Value
	switch val := ptr.(type) {
	case *string:
		value = (*stringValue)(val)
	case *bool:
		value = (*boolValue)(val)
	case *int:
		value = (*intValue)(val)
	case *int8:
		value = (*int8Value)(val)
	case *int16:
		value = (*int16Value)(val)
	case *int32:
		value = (*int32Value)(val)
	case *int64:
		value = (*int64Value)(val)
	case *uint:
		value = (*uintValue)(val)
	case *uint8:
		value = (*uint8Value)(val)
	case *uint16:
		value = (*uint16Value)(val)
	case *uint32:
		value = (*uint32Value)(val)
	case *uint64:
		value = (*uint64Value)(val)
	case *float64:
		value = (*float64Value)(val)
	case *float32:
		value = (*float32Value)(val)
	case *time.Duration:
		value = (*durationValue)(val)
	case Value:
		value = val
	default:
		rval := reflect.ValueOf(ptr)
		if rval.Kind() != reflect.Ptr {
			return nil, errors.New("ptr is not a pointer")
		}
		rind := reflect.Indirect(rval)
		var iValue = reflect.TypeOf(&value).Elem()
		if rind.Type().Implements(iValue) && rind.CanInterface() {
			value, _ = rind.Interface().(Value)
			return value, nil
		}
		return nil, errors.New("type " + rval.Type().String() + " can not be used as a Value")
	}
	return value, nil
}

type boolValue bool

func (p *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*p = boolValue(v)
	return err
}
func (p *boolValue) Get() interface{} { return *(*bool)(p) }
func (p *boolValue) String() string   { return strconv.FormatBool(bool(*p)) }
func (b *boolValue) IsBoolFlag() bool { return true }

type stringValue string

func (p *stringValue) Set(s string) error {
	*p = stringValue(s)
	return nil
}
func (p *stringValue) Get() interface{} { return p.String() }
func (p *stringValue) String() string   { return *(*string)(p) }

type byteValue byte

func (p *byteValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 8)
	*p = byteValue(v)
	return err
}
func (p *byteValue) Get() interface{} { return byte(*p) }
func (p *byteValue) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type uintValue uint

func (p *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 0)
	*p = uintValue(v)
	return err
}
func (p *uintValue) Get() interface{} { return uint(*p) }
func (p *uintValue) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type uint8Value uint8

func (p *uint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 8)
	*p = uint8Value(v)
	return err
}
func (p *uint8Value) Get() interface{} { return uint8(*p) }
func (p *uint8Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type uint16Value uint16

func (p *uint16Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 16)
	*p = uint16Value(v)
	return err
}
func (p *uint16Value) Get() interface{} { return uint16(*p) }
func (p *uint16Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type uint32Value uint32

func (p *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 32)
	*p = uint32Value(v)
	return err
}
func (p *uint32Value) Get() interface{} { return uint32(*p) }
func (p *uint32Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type uint64Value uint64

func (p *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*p = uint64Value(v)
	return err
}
func (p *uint64Value) Get() interface{} { return uint64(*p) }
func (p *uint64Value) String() string   { return strconv.FormatUint(uint64(*p), 10) }

type intValue int

func (p *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 0)
	*p = intValue(v)
	return err
}
func (p *intValue) Get() interface{} { return int(*p) }
func (p *intValue) String() string   { return strconv.FormatInt(int64(*p), 10) }

type int8Value int8

func (p *int8Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 8)
	*p = int8Value(v)
	return err
}
func (p *int8Value) Get() interface{} { return int8(*p) }
func (p *int8Value) String() string   { return strconv.FormatInt(int64(*p), 10) }

type int16Value int16

func (p *int16Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 16)
	*p = int16Value(v)
	return err
}
func (p *int16Value) Get() interface{} { return int16(*p) }
func (p *int16Value) String() string   { return strconv.FormatInt(int64(*p), 10) }

type int32Value int32

func (p *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	*p = int32Value(v)
	return err
}
func (p *int32Value) Get() interface{} { return int32(*p) }
func (p *int32Value) String() string   { return strconv.FormatInt(int64(*p), 10) }

type int64Value int64

func (p *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*p = int64Value(v)
	return err
}
func (p *int64Value) Get() interface{} { return int64(*p) }
func (p *int64Value) String() string   { return strconv.FormatInt(int64(*p), 10) }

type float32Value float32

func (p *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	*p = float32Value(v)
	return err
}
func (p *float32Value) Get() interface{} { return float32(*p) }
func (p *float32Value) String() string   { return strconv.FormatFloat(float64(*p), 'g', -1, 32) }

type float64Value float64

func (p *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*p = float64Value(v)
	return err
}
func (p *float64Value) Get() interface{} { return float64(*p) }
func (p *float64Value) String() string   { return strconv.FormatFloat(float64(*p), 'g', -1, 64) }

type durationValue time.Duration

func (p *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*p = durationValue(v)
	return err
}
func (p *durationValue) Get() interface{} { return time.Duration(*p) }
func (p *durationValue) String() string   { return p.String() }
