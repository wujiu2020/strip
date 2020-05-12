package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Context(t *testing.T) {
	type A struct {
		Value string
	}
	type B struct {
		Value string
	}
	var a A
	a.Value = "value"
	assert.Equal(t, key("pkg.jimu.io/libs/util:A:"), ctxValueToKey(&a))
	p := &a
	assert.Equal(t, key("pkg.jimu.io/libs/util:A:"), ctxValueToKey(&p))
	assert.Equal(t, key("pkg.jimu.io/libs/util:A:name"), ctxValueToKey(&p, "name"))
	assert.Panics(t, func() {
		ctxValueToKey(a)
	})

	assert.Panics(t, func() {
		q := &p
		ctxValueToKey(&q)
	})

	// not support non-pointer value
	assert.Panics(t, func() {
		ctxValueToKey(a)
	})

	// not support golang base type
	assert.Panics(t, func() {
		var a int
		ctxValueToKey(&a)
	})

	ctx := CtxWithValue(context.Background(), &a)

	// not found value
	var e *B
	err := CtxFindValue(ctx, &e)
	assert.Equal(t, CtxValueNotFound, err)

	// invalid value
	var iv *A
	assert.Panics(t, func() {
		CtxFindValue(ctx, iv)
	})

	// nil to value
	nilCtx := CtxWithValue(context.Background(), iv)
	var nv *A
	err = CtxFindValue(nilCtx, &nv)
	assert.NoError(t, err)
	assert.Equal(t, true, nv == nil)

	// *value to *value
	var b A
	err = CtxFindValue(ctx, &b)
	assert.NoError(t, err)
	assert.Equal(t, a, b)

	// *value to **value
	var c *A
	err = CtxFindValue(ctx, &c)
	assert.NoError(t, err)
	assert.Equal(t, a, *c)

	// **value to **value
	var pr *A
	ctx = CtxWithValue(context.Background(), &p)
	err = CtxFindValue(ctx, &pr)
	assert.NoError(t, err)
	assert.Equal(t, a, *pr)

	// **value to *value
	var pb A
	ctx = CtxWithValue(context.Background(), &p)
	err = CtxFindValue(ctx, &pb)
	assert.NoError(t, err)
	assert.Equal(t, a, pb)

	type E interface {
		Error() string
	}
	var er1 E = errors.New("error")
	var er2 E = errors.New("no_error")
	ctx = CtxWithValue(context.Background(), &er1)
	err = CtxFindValue(ctx, &er2)
	assert.NoError(t, err)
	assert.Equal(t, er1, er2)
}
