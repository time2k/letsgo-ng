package letsgo

import (
	"context"
	"errors"
	"time"
)

//ContextSet 结构体
type contextOne struct {
	CTX           context.Context
	CTXCancelFunc context.CancelFunc
}

type contextSet map[string]contextOne

//newContextSet 返回一个ContextSet结构体指针
func newContextSet() contextSet {
	cs := make(contextSet, 0)
	return cs
}

//新建一个ctx
func (ct contextSet) New(ctxtype string, ctxname string, parentname string, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	//存在同名
	if _, ok := ct[ctxname]; ok {
		return nil, nil, errors.New("[error]contextSet New ctxname exists!" + ctxname)
	}
	var parent context.Context
	if parentname == "" {
		parent = context.Background()
	} else {
		var err error
		parent, _, err = ct.Get(ctxname)
		if err != nil {
			return nil, nil, errors.New("[error]contextSet New parentname err:" + parentname + " " + err.Error())
		}
	}
	co := contextOne{}
	switch ctxtype {
	case "withcancel":
		co.CTX, co.CTXCancelFunc = context.WithCancel(parent)
	case "withtimeout":
		co.CTX, co.CTXCancelFunc = context.WithTimeout(parent, timeout)
	default:
		return nil, nil, errors.New("[error]contextSet New ctxtype not support:" + ctxtype)
	}
	ct[ctxname] = co

	return co.CTX, co.CTXCancelFunc, nil
}

//获取一个ctx
func (ct contextSet) Get(ctxname string) (context.Context, context.CancelFunc, error) {
	if _, ok := ct[ctxname]; !ok {
		return nil, nil, errors.New("[error]contextSet Get none-exists ctxname:" + ctxname)
	}
	gt := ct[ctxname]
	return gt.CTX, gt.CTXCancelFunc, nil
}

//依次cancel所有ctx
func (ct contextSet) CancelAll() {
	for _, ctxo := range ct {
		ctxo.CTXCancelFunc()
	}
	return
}
