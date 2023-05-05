package util

import (
	"context"
	"net/http"
)

type CtxKey string

func GetCtxVal(req *http.Request, ck CtxKey)interface{} {
	ctx := req.Context()
	return ctx.Value(ck)
}

func SetCtxKeyVal(r *http.Request, ck CtxKey, val interface{}) *http.Request {
	ctx := context.WithValue(r.Context(),ck,val)
	return r.WithContext(ctx)
}