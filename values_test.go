package envflag

import (
	"testing"
	"time"
)

func TestValueOfErrors(t *testing.T) {
	var err error
	_, err = ValueOf(nil)
	if err == nil {
		t.Errorf("ValueOf(nil) must cause an error")
	}

	var bs []byte
	_, err = ValueOf(&bs)
	if err == nil {
		t.Errorf("ValueOf(unsupported type) must cause an error")
	}

	var b byte
	_, err = ValueOf(b)
	if err == nil {
		t.Errorf("ValueOf(non pointer to upported type) must cause an error")
	}
}

func TestValueOfIntBounds(t *testing.T) {
	type inttest struct {
		ptr     interface{}
		lowin   string
		highin  string
		lowout  string
		highout string
	}
	var (
		u   uint
		u8  uint8 // also byte
		u16 uint16
		u32 uint32
		u64 uint64
		i   int
		i8  int8
		i16 int16
		i32 int32 // also rune
		i64 int64
	)
	tests := []inttest{
		{&u, "0", "18446744073709551615", "-1", "18446744073709551616"},
		{&u8, "0", "255", "-1", "256"},
		{&u16, "0", "65535", "-1", "65536"},
		{&u32, "0", "4294967295", "-1", "4294967296"},
		{&u64, "0", "18446744073709551615", "-1", "18446744073709551616"},
		{&i, "-9223372036854775808", "9223372036854775807", "-9223372036854775809", "9223372036854775808"},
		{&i8, "-128", "127", "-129", "128"},
		{&i16, "-32768", "32767", "-32769", "32768"},
		{&i32, "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{&i64, "-9223372036854775808", "9223372036854775807", "-9223372036854775809", "9223372036854775808"},
	}
	for _, test := range tests {
		ptr := test.ptr
		v, err := ValueOf(ptr)
		if err != nil {
			t.Fatalf("ValueOf(%T) must not cause an error: %s", ptr, err)
		}
		// check setting Value to values near boundaries
		for _, valid := range []string{test.lowin, test.highin} {
			err = v.Set(valid)
			if err != nil {
				t.Errorf("could not set %T (%s) to valid %s", ptr, v, valid)
			}
			str := v.String()
			if str != valid {
				t.Errorf("%T: expected %s, got %s", ptr, valid, str)
			}
		}
		for _, invalid := range []string{test.lowout, test.highout} {
			before := v.String()
			err = v.Set(invalid)
			if err == nil {
				t.Errorf("could set %T to invalid %s", ptr, invalid)
			}
			str := v.String()
			if str != before {
				t.Errorf("%T: expected %s, got %s", ptr, before, str)
			}
		}
	}
}

func TestValueOfCustomValue(t *testing.T) {
	var v noValue
	var err error
	_, err = ValueOf(v)
	if err != nil {
		t.Errorf("ValueOf(custom Value) must not cause an error")
	}
	_, err = ValueOf(&v)
	if err != nil {
		t.Errorf("ValueOf(*custom Value) must not cause an error")
	}

}

type noValue struct{}

func (p noValue) Set(s string) error { return nil }
func (p noValue) Get() interface{}   { return nil }
func (p noValue) String() string     { return "" }

func TestValueOfFloats(t *testing.T) {
	// For detailed tests, see http://golang.org/src/pkg/strconv/atof_test.go
	// This only checks boundaries and special values
	type floattest struct {
		val  string
		errs bool
	}
	var (
		f32 float32
		f64 float64
	)
	for ptr, tests := range map[interface{}][]floattest{
		&f32: {
			{"3.4028235e+38", false},
			{"-3.4028235e+38", false},
			{"3.4028236e+38", true},
			{"-3.4028236e+38", true},
			{"+Inf", false},
			{"-Inf", false},
			{"NaN", false},
		},
		&f64: {
			{"1.7976931348623157e+308", false},
			{"-1.7976931348623157e+308", false},
			{"1.797693134862315808e+308", true},
			{"-1.797693134862315808e+308", true},
			{"+Inf", false},
			{"-Inf", false},
			{"NaN", false},
		},
	} {
		v, err := ValueOf(ptr)
		if err != nil {
			t.Fatalf("ValueOf(%T) must not cause an error: %s", ptr, err)
		}
		for _, test := range tests {
			before := v.String()
			err = v.Set(test.val)
			if test.errs {
				if err == nil {
					t.Errorf("could set %T (%s, expected %s) to invalid %s", ptr, v, before, test.val)
				}
			} else {
				if err != nil || v.String() != test.val {
					t.Errorf("could not set %T (%s) to valid %s", ptr, v, test.val)
				}
			}
		}
	}
}

func TestValueOfBool(t *testing.T) {
	// For detailed tests, see http://golang.org/src/pkg/strconv/atob_test.go
	b := true
	var ptr interface{} = &b
	v, err := ValueOf(ptr)
	if err != nil {
		t.Fatalf("ValueOf(%T) must not cause an error: %s", ptr, err)
	}
	expected := "true"
	str := v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
	expected = "false"
	err = v.Set(expected)
	str = v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
}

func TestValueOfString(t *testing.T) {
	s := "abc"
	var ptr interface{} = &s
	v, err := ValueOf(ptr)
	if err != nil {
		t.Fatalf("ValueOf(%T) must not cause an error: %s", ptr, err)
	}
	expected := s
	str := v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
	expected = "Hello, 世界"
	err = v.Set(expected)
	str = v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
}

func TestValueOfDuration(t *testing.T) {
	s := "1s"
	d, _ := time.ParseDuration(s)
	var ptr interface{} = &d
	v, err := ValueOf(ptr)
	if err != nil {
		t.Fatalf("ValueOf(%T) must not cause an error: %s", ptr, err)
	}
	expected := s
	str := v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
	expected = "-3h2m1s"
	err = v.Set(expected)
	str = v.String()
	if str != expected {
		t.Errorf("%T: expected %s, got %s", ptr, expected, str)
	}
}
