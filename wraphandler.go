package querybus

import (
	"context"
	"reflect"
)

// WrapHandler is a generic wrapper for query handlers that calls the provided handler by reflection
func WrapHandler(handler interface{}) QueryHandler {
	handlerFunc := reflect.ValueOf(handler)
	return QueryHandlerFunc(func(ctx context.Context, query interface{}) (interface{}, error) {
		results := handlerFunc.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(query),
		})

		if len(results) >= 2 {
			if results[1].IsNil() {
				return results[0].Interface(), nil
			}
			return results[0].Interface(), results[1].Interface().(error)
		}
		return nil, nil
	})
}
