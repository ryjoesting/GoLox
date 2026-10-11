package main

type Environment struct {
	values map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{values: make(map[string]any)}
}

func (environment *Environment) Define(name string, value any) {
	environment.values[name] = value
}

func (environment *Environment) Assign(name *Token, value any) error {
	if _, ok := environment.values[name.Lexeme]; !ok {
		return RuntimeError{
			Token:   name,
			Message: "Undefined variable '" + name.Lexeme + "'.",
		}
	}
	environment.values[name.Lexeme] = value
	return nil
}

func (environment *Environment) Get(name *Token) (any, error) {
	value, ok := environment.values[name.Lexeme]
	if !ok {
		return nil, RuntimeError{
			Token:   name,
			Message: "Undefined variable '" + name.Lexeme + "'.",
		}
	}
	return value, nil
}
