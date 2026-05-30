package di

import (
	"fmt"
	"reflect"
)

type Provider func(*Container) error

type Container struct {
	values map[reflect.Type]reflect.Value
}

func New() *Container {
	return &Container{
		values: make(map[reflect.Type]reflect.Value),
	}
}

func (c *Container) Provide(constructor any) error {
	value := reflect.ValueOf(constructor)
	if value.Kind() != reflect.Func {
		return fmt.Errorf("provide dependency: constructor must be a function")
	}

	args, err := c.resolveArgs(value.Type())
	if err != nil {
		return fmt.Errorf("provide dependency: %w", err)
	}

	results := value.Call(args)
	if len(results) == 0 {
		return fmt.Errorf("provide dependency: constructor must return at least one value")
	}

	if last := results[len(results)-1]; isError(last.Type()) {
		if !last.IsNil() {
			return fmt.Errorf("provide dependency: %w", last.Interface().(error))
		}
		results = results[:len(results)-1]
	}
	if len(results) == 0 {
		return fmt.Errorf("provide dependency: constructor returned only an error")
	}

	for _, result := range results {
		c.values[result.Type()] = result
	}
	return nil
}

func (c *Container) Invoke(function any) error {
	value := reflect.ValueOf(function)
	if value.Kind() != reflect.Func {
		return fmt.Errorf("invoke dependency: target must be a function")
	}

	args, err := c.resolveArgs(value.Type())
	if err != nil {
		return fmt.Errorf("invoke dependency: %w", err)
	}

	results := value.Call(args)
	if len(results) == 0 {
		return nil
	}

	last := results[len(results)-1]
	if isError(last.Type()) && !last.IsNil() {
		return fmt.Errorf("invoke dependency: %w", last.Interface().(error))
	}
	return nil
}

func (c *Container) resolveArgs(functionType reflect.Type) ([]reflect.Value, error) {
	args := make([]reflect.Value, 0, functionType.NumIn())
	for i := 0; i < functionType.NumIn(); i++ {
		argType := functionType.In(i)
		value, ok := c.values[argType]
		if !ok {
			return nil, fmt.Errorf("missing dependency %s", argType)
		}
		args = append(args, value)
	}
	return args, nil
}

func isError(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
