package envflag

import (
	"math/big"
	"testing"
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

func getIntBounds(max string, signed bool) (lowout, lowin, highin, highout string, ok bool) {
	// use math/big to also generate outer bounds of int64 / uint64
	m1 := big.NewInt(-1)
	p1 := big.NewInt(1)
	tmp := big.NewInt(0)
	highin = max
	lowin = "0"
	if signed {
		tmp, ok = tmp.SetString("-"+highin, 10)
		if !ok {
			return
		}
		tmp = tmp.Add(tmp, m1)
		lowin = tmp.String()
	}
	lowout = tmp.Add(tmp, m1).String()
	tmp, ok = tmp.SetString(highin, 10)
	if !ok {
		return
	}
	highout = tmp.Add(tmp, p1).String()
	return
}

func TestValueOfIntBounds(t *testing.T) {
	// set variables to upper bound
	var (
		b   byte   = ^byte(0)
		u   uint   = ^uint(0)
		u8  uint8  = ^uint8(0)
		u16 uint16 = ^uint16(0)
		u32 uint32 = ^uint32(0)
		u64 uint64 = ^uint64(0)
		i   int    = int(u >> 1)
		i8  int8   = int8(u8 >> 1)
		i16 int16  = int16(u16 >> 1)
		i32 int32  = int32(u32 >> 1)
		i64 int64  = int64(u64 >> 1)
	)
	for _, ptr := range []interface{}{
		&b,
		&u, &u8, &u16, &u32, &u64,
		&i, &i8, &i16, &i32, &i64,
	} {
		v, err := ValueOf(ptr)
		if err != nil {
			t.Errorf("ValueOf(%T) must not cause an error: %s", ptr, err)
		}
		// use the starting value as maximum, get strings in and out of bounds
		signed := false
		switch ptr.(type) {
		case *int, *int8, *int16, *int32, *int64:
			signed = true
		}
		lo, li, hi, ho, ok := getIntBounds(v.String(), signed)
		if !ok {
			t.Errorf("could not convert bounds")
		}

		// DEBUG for checking boundaries per type:
		// fmt.Printf("%T:\n\t%s\n\t%s\n\t%s\n\t%s\n", ptr, lo, li, hi, ho)

		// check setting Value to values near boundaries
		for _, valid := range []string{li, hi} {
			err = v.Set(valid)
			if err != nil {
				t.Errorf("could not set %T to valid %s", ptr, valid)
			}
			str := v.String()
			if str != valid {
				t.Errorf("%T: expected %s, got %s", ptr, valid, str)
			}
		}
		for _, invalid := range []string{lo, ho} {
			before := v.String()
			err = v.Set(invalid)
			if err == nil {
				t.Errorf("could set %T to invalid %s", ptr, invalid)
			}
			str := v.String()
			if str == before {
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

// TODO missing float32, float64, bool, string, time.Duration
