package params

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPositiveBindValuesToStruct(t *testing.T) {
	type aliasString string

	type embeddedStruct struct {
		Embedded string `param:"embedded"`
	}

	type embededPointer struct {
		Value string `param:"value"`
	}

	type input struct {
		// embeded field
		embeddedStruct

		// embeded struct pointer
		*embededPointer

		// builtin type
		Int64Min  int64   `param:"int64min"`
		Int64Max  int64   `param:"int64max"`
		Uint64Min uint64  `param:"uint64min"`
		Uint64Max uint64  `param:"uint64max"`
		Float64   float64 `param:"float64"`
		String    string  `param:"string"`
		BoolTrue  bool    `param:"booltrue"`
		BoolOn    bool    `param:"boolon"`
		Bool1     bool    `param:"bool1"`
		BoolFalse bool    `param:"boolfalse"`

		// time
		Date        time.Time `param:"date"`
		DateTime    time.Time `param:"datetime"`
		TimeUnix    time.Time `param:"timeunix"`
		TimeRFC3339 time.Time `param:"timerfc3339"`

		// alias type
		AliasString aliasString `param:"aliasstring"`

		// map
		Map map[string]string `param:"map"`

		// slice
		Slice []string `param:"slice"`
	}

	paramsIn := &Params{
		Values: url.Values{
			"embedded": []string{"helloworld"},

			"int64min":  []string{"-9223372036854775808"},
			"int64max":  []string{"9223372036854775807"},
			"uint64min": []string{"0"},
			"uint64max": []string{"18446744073709551615"},
			"float64":   []string{"3.1415926"},
			"string":    []string{"abcdefg"},
			"booltrue":  []string{"true"},
			"boolon":    []string{"on"},
			"bool1":     []string{"1"},
			"boolfalse": []string{"false"},

			"date":        []string{"2006-01-02"},
			"datetime":    []string{"2006-01-02 15:04:05"},
			"timeunix":    []string{"1136214245"},
			"timerfc3339": []string{"2006-01-02T15:04:05Z08:00"},

			"aliasstring": []string{"alphago"},
			"map[foo]":    []string{"bar"},
			"slice[]":     []string{"one", "two"},
		},
	}

	var structActual input

	date, _ := time.ParseInLocation("2006-01-02", "2006-01-02", time.Local)
	datetime, _ := time.ParseInLocation("2006-01-02 15:04:05", "2006-01-02 15:04:05", time.Local)
	timeunix := time.Unix(1136214245, 0)
	timerfc3339, _ := time.ParseInLocation(time.RFC3339, "2006-01-02T15:04:05Z08:00", time.Local)

	structExpected := input{
		embeddedStruct: embeddedStruct{Embedded: "helloworld"},

		Int64Min:  -9223372036854775808,
		Int64Max:  9223372036854775807,
		Uint64Min: 0,
		Uint64Max: 18446744073709551615,
		Float64:   3.1415926,
		String:    "abcdefg",
		BoolTrue:  true,
		BoolOn:    true,
		Bool1:     true,
		BoolFalse: false,

		Date:        date,
		DateTime:    datetime,
		TimeUnix:    timeunix,
		TimeRFC3339: timerfc3339,

		AliasString: aliasString("alphago"),

		Map: map[string]string{"foo": "bar"},

		Slice: []string{"one", "two"},
	}

	paramsIn.BindValuesToStruct(&structActual)

	assert.Equal(t, structExpected.Embedded, structActual.Embedded, "embeded testing")
	assert.Equal(t, structExpected.Int64Min, structActual.Int64Min, "int64 min testing")
	assert.Equal(t, structExpected.Int64Max, structActual.Int64Max, "int64 max testing")
	assert.Equal(t, structExpected.Uint64Min, structActual.Uint64Min, "uint64 min testing")
	assert.Equal(t, structExpected.Uint64Max, structActual.Uint64Max, "uint64 max testing")
	assert.Equal(t, structExpected.Float64, structActual.Float64, "float64 testing")
	assert.Equal(t, structExpected.String, structActual.String, "string testing")
	assert.Equal(t, structExpected.BoolTrue, structActual.BoolTrue, "bool 'true' testing")
	assert.Equal(t, structExpected.BoolOn, structActual.BoolOn, "bool 'on' testing")
	assert.Equal(t, structExpected.Bool1, structActual.Bool1, "bool '1' testing")
	assert.Equal(t, structExpected.BoolFalse, structActual.BoolFalse, "bool 'false' testing")
	assert.Equal(t, structExpected.Date.Unix(), structActual.Date.Unix(), "date testing")
	assert.Equal(t, structExpected.DateTime.Unix(), structActual.DateTime.Unix(), "datetime testing")
	assert.Equal(t, structExpected.TimeUnix.Unix(), structActual.TimeUnix.Unix(), "unix time testing")
	assert.Equal(t, structExpected.TimeRFC3339.Unix(), structActual.TimeRFC3339.Unix(), "RFC3339 time testing")
	assert.Equal(t, structExpected.AliasString, structActual.AliasString, "alias testing")
	assert.Equal(t, structExpected.Map["foo"], structActual.Map["foo"], "map testing")
	assert.Equal(t, structExpected.Slice, structActual.Slice, "slice testing")

	assert.Equal(t, true, structActual.embededPointer == nil)
	paramsIn.Values.Set("value", "value")
	paramsIn.BindValuesToStruct(&structActual)
	assert.Equal(t, true, structActual.embededPointer != nil)
	if structActual.embededPointer != nil {
		assert.Equal(t, true, structActual.embededPointer.Value == "value")
	}
}
