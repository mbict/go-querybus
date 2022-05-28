package querybus

import (
	"context"
	"errors"
	"reflect"
)

var ErrUnhandledQuery = errors.New("unhandled query")
var ErrConvertQueryCommand = errors.New("unable to convert query command")
var ErrDuplicateQueryHandler = errors.New("there is already a query handler registered for this query")

type Query interface {
	QueryName() string
}

type QueryHandler interface {
	Handle(ctx context.Context, command interface{}) (interface{}, error)
}

type QueryHandlerFunc func(ctx context.Context, query interface{}) (interface{}, error)

func (h QueryHandlerFunc) Handle(ctx context.Context, command interface{}) (interface{}, error) {
	return h(ctx, command)
}

func RegisterHandler[Q Query, O any](querybus QueryBus, handler func(context.Context, Q) (O, error)) error {
	return querybus.Register(new(Q), H(handler))
}

func H[Q Query, O any](handler func(context.Context, Q) (O, error)) QueryHandler {
	return QueryHandlerFunc(func(ctx context.Context, query interface{}) (interface{}, error) {
		q, ok := query.(Q)
		if !ok {
			return nil, ErrConvertQueryCommand
		}
		return handler(ctx, q)
	})
}

type QueryBus interface {
	QueryHandler
	Register(interface{}, QueryHandler) error
}

func New() QueryBus {
	return &queryBus{
		handlers: make(map[string]QueryHandler),
	}
}

type queryBus struct {
	handlers map[string]QueryHandler
}

func (cb *queryBus) Register(command interface{}, handler QueryHandler) error {
	commandName := resolveQueryName(command)
	if _, ok := cb.handlers[commandName]; ok {
		return ErrDuplicateQueryHandler
	}
	cb.handlers[commandName] = handler
	return nil
}

func (cb *queryBus) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	if h, ok := cb.handlers[resolveQueryName(query)]; ok {
		return h.Handle(ctx, query)
	}
	return nil, ErrUnhandledQuery
}

func resolveQueryName(query interface{}) string {
	if c, ok := query.(Query); ok {
		return c.QueryName()
	}

	t := reflect.TypeOf(query)
	if t.Kind() == reflect.Ptr {
		return t.Elem().PkgPath() + "/*" + t.Elem().Name()
	}
	return t.PkgPath() + "/" + t.Name()
}
