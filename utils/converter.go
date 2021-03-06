package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
)

// convert string to specify type
type StrTo string

func (f StrTo) Bool() (bool, error) {
	if f == "on" {
		return true, nil
	}
	return strconv.ParseBool(f.String())
}

func (f StrTo) Float32() (float32, error) {
	v, err := strconv.ParseFloat(f.String(), 32)
	return float32(v), err
}

func (f StrTo) Float64() (float64, error) {
	return strconv.ParseFloat(f.String(), 64)
}

func (f StrTo) Int() (int, error) {
	v, err := strconv.ParseInt(f.String(), 10, 64)
	return int(v), err
}

func (f StrTo) Int8() (int8, error) {
	v, err := strconv.ParseInt(f.String(), 10, 8)
	return int8(v), err
}

func (f StrTo) Int16() (int16, error) {
	v, err := strconv.ParseInt(f.String(), 10, 16)
	return int16(v), err
}

func (f StrTo) Int32() (int32, error) {
	v, err := strconv.ParseInt(f.String(), 10, 32)
	return int32(v), err
}

func (f StrTo) Int64() (int64, error) {
	v, err := strconv.ParseInt(f.String(), 10, 64)
	return int64(v), err
}

func (f StrTo) Uint() (uint, error) {
	v, err := strconv.ParseUint(f.String(), 10, 64)
	return uint(v), err
}

func (f StrTo) Uint8() (uint8, error) {
	v, err := strconv.ParseUint(f.String(), 10, 8)
	return uint8(v), err
}

func (f StrTo) Uint16() (uint16, error) {
	v, err := strconv.ParseUint(f.String(), 10, 16)
	return uint16(v), err
}

func (f StrTo) Uint32() (uint32, error) {
	v, err := strconv.ParseUint(f.String(), 10, 32)
	return uint32(v), err
}

func (f StrTo) Uint64() (uint64, error) {
	v, err := strconv.ParseUint(f.String(), 10, 64)
	return uint64(v), err
}

func (f StrTo) MustBool() bool {
	v, _ := f.Bool()
	return v
}

func (f StrTo) MustFloat32() float32 {
	v, _ := f.Float32()
	return v
}

func (f StrTo) MustFloat64() float64 {
	v, _ := f.Float64()
	return v
}

func (f StrTo) MustInt() int {
	v, _ := f.Int()
	return v
}

func (f StrTo) MustInt8() int8 {
	v, _ := f.Int8()
	return v
}

func (f StrTo) MustInt16() int16 {
	v, _ := f.Int16()
	return v
}

func (f StrTo) MustInt32() int32 {
	v, _ := f.Int32()
	return v
}

func (f StrTo) MustInt64() int64 {
	v, _ := f.Int64()
	return v
}

func (f StrTo) MustUint() uint {
	v, _ := f.Uint()
	return v
}

func (f StrTo) MustUint8() uint8 {
	v, _ := f.Uint8()
	return v
}

func (f StrTo) MustUint16() uint16 {
	v, _ := f.Uint16()
	return v
}

func (f StrTo) MustUint32() uint32 {
	v, _ := f.Uint32()
	return v
}

func (f StrTo) MustUint64() uint64 {
	v, _ := f.Uint64()
	return v
}

func (f StrTo) Bytes() []byte {
	return []byte(f)
}

func (f StrTo) Md5() string {
	h := md5.New()
	h.Write(f.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}

func (f StrTo) String() string {
	return string(f)
}

// convert any type to string
func ToStr(value interface{}, args ...int) (s string) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Bool:
		s = strconv.FormatBool(val.Bool())
	case reflect.Float32:
		s = strconv.FormatFloat(val.Float(), 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 32))
	case reflect.Float64:
		s = strconv.FormatFloat(val.Float(), 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 64))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s = strconv.FormatInt(val.Int(), argInt(args).Get(0, 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s = strconv.FormatUint(val.Uint(), argInt(args).Get(0, 10))
	case reflect.String:
		s = val.String()
	default:
		if v, ok := value.([]byte); ok {
			// TODO should use reflect rewrite
			s = string(v)
		} else {
			s = fmt.Sprintf("%#v", v)
		}
	}
	return s
}

// convert any numeric value to int64
func ToInt64(value interface{}) (d int64, err error) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d = val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d = int64(val.Uint())
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	return
}

type argString []string

func (a argString) Get(i int, args ...string) (r string) {
	if i >= 0 && i < len(a) {
		r = a[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}

type argInt []int

func (a argInt) Get(i int, args ...int) (r int) {
	if i >= 0 && i < len(a) {
		r = a[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}

type argAny []interface{}

func (a argAny) Get(i int, args ...interface{}) (r interface{}) {
	if i >= 0 && i < len(a) {
		r = a[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}
