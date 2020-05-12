package utils

import (
	"context"
	"fmt"
	"reflect"
)

var (
	CtxValueNotFound = fmt.Errorf("value not found in context")
)

type key string

func CtxWithValue(ctx context.Context, value interface{}, extraName ...string) context.Context {
	k := ctxValueToKey(value, extraName...)
	return context.WithValue(ctx, k, value)
}

func CtxFindValue(ctx context.Context, value interface{}, extraName ...string) error {
	k := ctxValueToKey(value, extraName...)
	inf := ctx.Value(k)
	tarElm := reflect.ValueOf(value).Elem()
	if !tarElm.CanSet() {
		panic(fmt.Errorf("wrong value: %q cannot set", tarElm))
	}
	if inf == nil {
		return CtxValueNotFound
	}
	val := reflect.ValueOf(inf)
	if val.IsNil() {
		return nil
	}
	elm := val.Elem()
	if tarElm.Kind() == reflect.Ptr && elm.Kind() != reflect.Ptr {
		tarElm.Set(elm.Addr())
	} else if tarElm.Kind() != reflect.Ptr && elm.Kind() == reflect.Ptr {
		tarElm.Set(elm.Elem())
	} else if tarElm.Kind() == reflect.Ptr && elm.Kind() == reflect.Ptr {
		tarElm.Set(elm)
	} else {
		tarElm.Set(elm)
	}
	return nil
}

func MustCtxFindValue(ctx context.Context, value interface{}, extraName ...string) {
	err := CtxFindValue(ctx, value, extraName...)
	if err != nil {
		panic(err)
	}
}

func ctxValueToKey(value interface{}, extraNames ...string) key {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr {
		panic("must use pointer")
	}
	elm := val.Type().Elem()
	if elm.Kind() == reflect.Ptr {
		elm = elm.Elem()
	}
	if elm.Kind() == reflect.Ptr {
		panic("not support ***pointer")
	}
	if elm.Kind().String() == elm.String() {
		panic("not support golang base type")
	}
	var extraName string
	if len(extraNames) > 0 {
		extraName = extraNames[0]
	}
	v := elm.PkgPath() + ":" + elm.Name() + ":" + extraName
	return key(v)
}
