package value

import (
	"fmt"
	"math"
	"testing"
	"time"
)

const maxStringLen = 128

func TestValueOfBool(t *testing.T) {
	// For detailed tests, see http://golang.org/src/pkg/strconv/atob_test.go
	b0 := false
	b1 := true
	v := testValidCases(t, &b0, b0, b1)

	type boolFlag interface {
		IsBoolFlag() bool
	}
	if bf, ok := v.(boolFlag); !(ok && bf.IsBoolFlag()) {
		t.Errorf("for bool, IsBoolFlag() must be true")
	}
}

func TestValueOfIntegerTypes(t *testing.T) {
	type inttest struct {
		// value and pointer to value
		val interface{}
		ptr interface{}
		// bounds
		lowin   string
		highin  string
		lowout  string
		highout string
	}
	var (
		u   uint   = 1
		u8  uint8  = 2 // also byte
		u16 uint16 = 3
		u32 uint32 = 4
		u64 uint64 = 5
		i   int    = -1
		i8  int8   = -2
		i16 int16  = -3
		i32 int32  = -4 // also rune
		i64 int64  = -5
	)
	tests := []inttest{
		{u, &u, "0", "18446744073709551615", "-1", "18446744073709551616"},
		{u8, &u8, "0", "255", "-1", "256"},
		{u16, &u16, "0", "65535", "-1", "65536"},
		{u32, &u32, "0", "4294967295", "-1", "4294967296"},
		{u64, &u64, "0", "18446744073709551615", "-1", "18446744073709551616"},
		{i, &i, "-9223372036854775808", "9223372036854775807", "-9223372036854775809", "9223372036854775808"},
		{i8, &i8, "-128", "127", "-129", "128"},
		{i16, &i16, "-32768", "32767", "-32769", "32768"},
		{i32, &i32, "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{i64, &i64, "-9223372036854775808", "9223372036854775807", "-9223372036854775809", "9223372036854775808"},
	}
	for _, test := range tests {
		ptr := test.ptr
		v, ok := ValueOf(ptr)
		if !ok {
			t.Fatalf("ValueOf([%T]) failed", ptr)
		}
		get, ok := v.(Value)
		if !ok {
			t.Fatalf("ValueOf([%T]) result must provide Get()", ptr)
		}
		if gv := get.Get(); gv != test.val {
			t.Errorf("Get() on ValueOf([%T]) result must return initial value %v, was %v", ptr, test.val, gv)
		}
		// check setting Value to values near boundaries
		for _, valid := range []string{test.lowin, test.highin} {
			if err := v.Set(valid); err != nil {
				t.Errorf("could not set %T (%s) to valid %s", ptr, v, valid)
			}
			str := v.String()
			if str != valid {
				t.Errorf("%T: expected %s, got %s", ptr, valid, str)
			}
		}
		for _, invalid := range []string{test.lowout, test.highout} {
			before := v.String()
			if err := v.Set(invalid); err == nil {
				t.Errorf("could set %T to invalid %s", ptr, invalid)
			}
			str := v.String()
			if str != before {
				t.Errorf("%T: expected %s, got %s", ptr, before, str)
			}
			if a := string(v.AppendTo(nil)); a != str {
				t.Errorf("%T.AppendTo(nil): expected %s, got %s", ptr, a, str)
			}
		}
	}
}

func TestValueOfFloat32(t *testing.T) {
	vals := []float32{
		0,
		math.SmallestNonzeroFloat32,
		3.4028235e+38,
		float32(math.Inf(1)),
	}
	use := make([]interface{}, 2*len(vals)-1)
	use[0] = vals[0]
	for i := 1; i < len(use); i += 2 {
		val := vals[i/2]
		use[i], use[i+1] = val, -val
	}
	v := testValidCases(t, &vals[0], use...)

	if err := v.Set("NaN"); err != nil {
		t.Errorf("could not set to NaN")
	} else if fn, ok := v.Get().(float32); !(ok && math.IsNaN(float64(fn))) {
		t.Errorf("could not retrieve NaN value")
	}

	// %e instead of %g because %g uses 64 bit precision for float32
	max := fmt.Sprintf("%e", math.MaxFloat32)
	err := v.Set(max)
	if err != nil {
		t.Errorf("could not set max value")
	}
	if str := v.String(); max != str {
		t.Errorf("max value string mismatch: expected %s, got %s", max, str)
	}

	if err := v.Set("3.4028236e+38"); err == nil {
		t.Errorf("expected a value on setting a too large value")
	}
}

func TestValueOfFloat64(t *testing.T) {
	vals := []float64{
		0,
		math.SmallestNonzeroFloat64,
		1.7976931348623157e+308,
		math.Inf(1),
	}
	use := make([]interface{}, 2*len(vals)-1)
	use[0] = vals[0]
	for i := 1; i < len(use); i += 2 {
		val := vals[i/2]
		use[i], use[i+1] = val, -val
	}
	v := testValidCases(t, &vals[0], use...)

	if err := v.Set("NaN"); err != nil {
		t.Errorf("could not set to NaN")
	} else if fn, ok := v.Get().(float64); !(ok && math.IsNaN(fn)) {
		t.Errorf("could not retrieve NaN value")
	}

	max := fmt.Sprintf("%g", math.MaxFloat64)
	err := v.Set(max)
	if err != nil {
		t.Errorf("could not set max value")
	}
	if str := v.String(); max != str {
		t.Errorf("max value string mismatch: expected %s, got %s", max, str)
	}

	if err := v.Set("1.797693134862315808e+308"); err == nil {
		t.Errorf("expected a value on setting a too large value")
	}
}

func TestValueOfString(t *testing.T) {
	s0 := ""
	s1 := "abc"
	s2 := " "
	s3 := "Hello, 世界"
	testValidCases(t, &s0, s0, s1, s2, s3)
	testValidCases(t, &s3, s0, s1, s2, s3)
}

func TestValueOfDuration(t *testing.T) {
	d0 := 1 * time.Second
	d1 := -(3*time.Hour + 2*time.Minute + 1*time.Second)
	testValidCases(t, &d0, d0, d1)
}

func TestValueOfBadArgs(t *testing.T) {
	if _, ko := ValueOf(nil); ko {
		t.Errorf("ValueOf(nil) must cause an error")
	}

	var b byte
	if _, ko := ValueOf(b); ko {
		t.Errorf("ValueOf([non pointer to supported builtin type]) must cause an error")
	}

	var bs []byte
	if _, ko := ValueOf(&bs); ko {
		t.Errorf("ValueOf([pointer to unsupported type]) must cause an error")
	}
}

// testValidCases retrieves a Value for v0ptr and runs basic tests on it.
func testValidCases(t *testing.T, v0ptr interface{}, values ...interface{}) Value {
	v, ok := ValueOf(v0ptr)
	if !ok {
		t.Fatalf("ValueOf([%T]) failed", v0ptr)
	}
	buf := [maxStringLen]byte{}
	for _, val := range values {
		s := fmt.Sprintf("%v", val)
		if err := v.Set(s); err != nil {
			t.Errorf("%T Set(%q) must not cause error, got %s", v0ptr, s, err)
		}
		if got := v.Get(); got != val {
			t.Errorf("%T Get(): expected %v, got %v", v0ptr, val, got)
		}
		if vs := v.String(); s != vs {
			t.Errorf("%T String(): expected %s, got %s", v0ptr, s, vs)
		}
		if vs := string(v.AppendTo(nil)); s != vs {
			t.Errorf("%T AppendTo(...): expected %s, got %s", v0ptr, s, vs)
		}
		vbytes := v.AppendTo(buf[:0])
		if vs := string(vbytes); s != vs {
			t.Errorf("%T AppendTo(...): expected %s, got %s", v0ptr, s, vs)
		}
		if len(vbytes) > 0 && &buf[0] != &vbytes[0] {
			t.Errorf("%T AppendTo(): fits in cap but does not use buffer (len %d, cap %d)", v0ptr, len(vbytes), cap(buf[:0]))
		}
	}
	return v
}
