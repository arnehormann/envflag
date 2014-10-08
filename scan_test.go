package envflag

import (
	"testing"
	"time"

	"github.com/confactor/envflag/value"
)

func testDeadline(t *testing.T, d time.Duration, test func(t *testing.T)) {
	terminated := make(chan struct{})
	go func(t *testing.T, test func(t *testing.T), terminated chan<- struct{}) {
		test(t)
		terminated <- struct{}{}
	}(t, test, terminated)
	timedout := time.After(d)
	select {
	case <-timedout:
		panic("test exceeded deadline")
	case <-terminated:
		close(terminated)
	}
}

func TestScanBuiltins(t *testing.T) {
	type Builtin struct {
		F00 bool
		F01 int
		F02 int8
		F03 int16
		F04 int32
		F05 int64
		F06 uint
		F07 uint8
		F08 uint16
		F09 uint32
		F0A uint64
		F0B float32
		F0C float64
		F0D string
		F0E time.Duration
	}
	v := Builtin{}
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
	type BuiltinPtrs struct {
		F00 *bool
		F01 *int
		F02 *int8
		F03 *int16
		F04 *int32
		F05 *int64
		F06 *uint
		F07 *uint8
		F08 *uint16
		F09 *uint32
		F0A *uint64
		F0B *float32
		F0C *float64
		F0D *string
		F0E *time.Duration
	}
	p := BuiltinPtrs{
		F00: &v.F00,
		F01: &v.F01,
		F02: &v.F02,
		F03: &v.F03,
		F04: &v.F04,
		F05: &v.F05,
		F06: &v.F06,
		F07: &v.F07,
		F08: &v.F08,
		F09: &v.F09,
		F0A: &v.F0A,
		F0B: &v.F0B,
		F0C: &v.F0C,
		F0D: &v.F0D,
		F0E: &v.F0E,
	}
	if _, err := ScanWarn(&p); err != nil {
		t.Error(err)
	}
}

func TestScanInvalid(t *testing.T) {
	type Invalid struct {
		C chan byte
	}
	v := Invalid{}
	if _, err := ScanWarn(&v); err != nil {
		warn, ok := err.(*ScanWarnings)
		if !ok || len(warn.Skipped) != 1 || warn.Skipped[0] != "C" {
			t.Error(err)
		}
	} else {
		t.Errorf("expected warnign on skipped parameter")
	}
}

func TestScanEmptyOk(t *testing.T) {
	type Empty struct{}
	v := Empty{}
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
}

func TestScanArray(t *testing.T) {
	v := struct {
		S [2]string
	}{}
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
}

func TestScanSlice(t *testing.T) {
	v := struct {
		S interface{}
	}{make([]string, 2)}
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
	v.S = make([]string, 0)
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
	v.S = nil
	if _, err := ScanWarn(&v); err == nil {
		t.Errorf("expected a warning on skipped value")
	}
}

func TestScanEmbedded(t *testing.T) {
	type Value struct {
		I int
	}
	type Outer struct {
		Value
	}
	v := Outer{}
	if _, err := ScanWarn(&v); err != nil {
		t.Error(err)
	}
}

func TestScanParameterDuplicate(t *testing.T) {
	v := struct {
		I interface{}
		J int
	}{}
	v.I = &v.J

	if _, err := ScanWarn(&v); err != nil {
		warn, ok := err.(*ScanWarnings)
		if !ok || len(warn.Duplicates) != 1 {
			t.Error(err)
		} else {
			dupes := warn.Duplicates[0]
			if !(len(dupes) == 2 && dupes[0] == "I" && dupes[1] == "J") {
				t.Error(err)
			}
		}
	} else {
		t.Errorf("expected warnign on duplicate parameter")
	}
}

func TestScanDuplicates(t *testing.T) {
	/*
		checkDuplicates := func(err error, numSkip int) [][]string {
			warn, ok := err.(*ScanWarnings)
			if !ok {
				t.Fatalf("error %s is not *ScanWarning but %T", err, err)
			}
			if len(warn.Skipped) != numSkip {
				t.Fatalf("expected %d skipped values, got %s", numSkip, err)
			}
			return warn.Duplicates
		}
	*/
	type Duplicates struct {
		A interface{}
		B interface{}
	}
	i := 0
	v, _ := value.ValueOf(&i)
	params := Duplicates{
		A: v,
	}
	if _, err := Scan(&params); err != nil {
		t.Error(err)
	}
	if _, err := ScanWarn(&params); err == nil {
		t.Errorf("expected a warning")
	}
	params.B = v
	if _, err := Scan(&params); err != nil {
		t.Error(err)
	}
	if _, err := ScanWarn(&params); err == nil {
		t.Errorf("expected a warning")
	}
	j := 0
	params.A = &i
	params.B = &j
	if _, err := Scan(&params); err != nil {
		t.Error(err)
	}
	if _, err := ScanWarn(&params); err != nil {
		t.Error(err)
	}

	type Inner struct{ S string }
	inner := Inner{}
	o := Duplicates{
		A: &inner,
		B: &inner,
	}
	if _, err := ScanWarn(&o); err == nil {
		t.Error(err)
	}
	inner0 := Inner{}
	o.B = &inner0
	if _, err := ScanWarn(&o); err != nil {
		t.Error(err)
	}
}

func TestScanSkipping(t *testing.T) {
	type Empty struct{}
	type EmbeddedPtrNil Empty
	type EmbeddedPtrNotNil Empty

	type Skipping struct {
		*EmbeddedPtrNil
		*EmbeddedPtrNotNil
		PtrNil      *int
		PtrNotNil   *int
		unexported  int
		Unsupported interface{}
		Supported   interface{}
	}

	i := 0
	nn := EmbeddedPtrNotNil{}

	v := Skipping{
		EmbeddedPtrNotNil: &nn,
		PtrNotNil:         &i,
		Unsupported:       make(chan byte),
		Supported:         []string{"idx0", "idx1", "idx2"},
	}

	skips := []string{
		"EmbeddedPtrNil",
		"PtrNil",
		"unexported",
		"Unsupported",
	}

	_, err := ScanWarn(&v)

	if err == nil {
		t.Fatalf("expected a warning")
	}

	warn, ok := err.(*ScanWarnings)
	if !ok {
		t.Error(err)
	}

	if len(warn.Duplicates) > 0 {
		t.Errorf("expected no duplicates: %s", warn)
	}

	falseNegative := []string{}
	for _, path := range warn.Skipped {
		found := false
		for i, max := 0, len(skips); i < max; i++ {
			if skips[i] == path {
				skips[i] = ""
				found = true
			}
		}
		if !found {
			falseNegative = append(falseNegative, path)
		}
	}

	falsePositive := []string{}
	for _, path := range skips {
		if path != "" {
			falsePositive = append(falsePositive, path)
		}
	}

	if len(falseNegative)+len(falsePositive) > 0 {
		t.Errorf("false negatives: %v; false positives: %v",
			falseNegative,
			falsePositive,
		)
	}

	// no specified message format to discourage parsing, just a smoke test
	if msg := warn.Error(); msg == "" {
		t.Fatalf("warning message expected")
	}
}

func TestScanTerminatesOnCycles(t *testing.T) {
	const deadline = 100 * time.Millisecond

	v := struct {
		V interface{}
	}{}

	testDeadline(t, deadline, func(t *testing.T) {
		v.V = &v
		ScanWarn(&v)
	})

	testDeadline(t, deadline, func(t *testing.T) {
		v.V = []interface{}{&v}
		ScanWarn(&v)
	})

	testDeadline(t, deadline, func(t *testing.T) {
		ptr := &v
		v.V = &ptr
		ScanWarn(&v)
	})

	testDeadline(t, deadline, func(t *testing.T) {
		inner := struct {
			A **int
			B interface{}
		}{}
		i := 0
		pi := &i
		// intermediary pointer
		inner.A = &pi
		inner.B = pi
		v.V = inner
		ScanWarn(&v)
	})
}

func TestScanErrors(t *testing.T) {
	for name, scan := range map[string]func(interface{}) (Module, error){
		"Scan":     Scan,
		"ScanWarn": ScanWarn,
	} {
		if _, err := scan(nil); err == nil {
			t.Errorf("%s: expected error when scanning nil", name)
		}

		x := 0
		if _, err := scan(x); err == nil {
			t.Errorf("%s: expected error when scanning non pointer", name)
		}

		if _, err := scan(&x); err == nil {
			t.Errorf("%s: expected error when scanning non struct pointer", name)
		}

		type V struct {
			I int
		}
		vptr := &V{}
		if _, err := scan(&vptr); err == nil {
			t.Errorf("%s: expected error when scanning **struct", name)
		}
	}
}

func TestScanWarningMsg(t *testing.T) {
	// ScanWarnings is not intended to be modified, just checking Error() here

	warn := &ScanWarnings{}

	dups := [][]string{
		{"D0a/D0a_", "D0b"},
		{"D1a", "D1b"},
	}
	skips := []string{
		"S0/S0_",
		"S1",
	}

	if msg := warn.Error(); msg != "" {
		t.Fatalf("empty message expected on no warnings")
	}

	warn.Duplicates = dups[:1]
	if msg := warn.Error(); msg == "" {
		t.Fatalf("message expected on warnings")
	}
	warn.Duplicates = dups
	if msg := warn.Error(); msg == "" {
		t.Fatalf("message expected on warnings")
	}

	warn.Duplicates = nil
	warn.Skipped = skips[:1]
	if msg := warn.Error(); msg == "" {
		t.Fatalf("message expected on warnings")
	}
	warn.Skipped = skips
	if msg := warn.Error(); msg == "" {
		t.Fatalf("message expected on warnings")
	}

	warn.Duplicates = dups
	if msg := warn.Error(); msg == "" {
		t.Fatalf("message expected on warnings")
	}
}

/*

type noValue struct{}

func (p noValue) Set(s string) error { return nil }
func (p noValue) Get() interface{}   { return nil }
func (p noValue) String() string     { return "" }

func TestValueOfCustomValue(t *testing.T) {
	var v noValue

	if _, ok := ValueOf(&v); !ok {
		t.Errorf("ValueOf([pointer to custom Value]) must not cause an error")
	}

	if _, ok := ValueOf(v); !ok {
		t.Errorf("ValueOf([non pointer custom Value]) must not cause an error")
	}
}
*/
