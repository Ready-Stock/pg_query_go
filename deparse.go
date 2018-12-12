package pg_query

import (
	pq "github.com/readystock/pg_query_go/nodes"
	"github.com/kataras/go-errors"
	"reflect"
)

func DeparseValue(aconst pq.A_Const) (interface{}, error) {
	switch c := aconst.Val.(type) {
	case pq.String:
		return c.Str, nil
	case pq.Integer:
		return c.Ival, nil
	case pq.Null:
		return nil, nil
	default:
		return nil, errors.New("cannot parse type %s").Format(reflect.TypeOf(c).Name())
	}
}
