package main

type Environment struct {
	values    map[string]any
	enclosing *Environment
}

func NewEnvironment(enclosing ...*Environment) *Environment {
	var parent *Environment
	if len(enclosing) > 0 {
		parent = enclosing[0]
	}
	return &Environment{values: make(map[string]any), enclosing: parent}
}

func (environment *Environment) Define(name string, value any) {
	environment.values[name] = value
}

func (environment *Environment) Assign(name *Token, value any) error {
	if _, ok := environment.values[name.Lexeme]; ok {
		environment.values[name.Lexeme] = value
		return nil
	}
	if environment.enclosing != nil {
		return environment.enclosing.Assign(name, value)
	}
	return undefinedVariableError(name)
}

func (environment *Environment) Get(name *Token) (any, error) {
	if value, ok := environment.values[name.Lexeme]; ok {
		return value, nil
	}
	if environment.enclosing != nil {
		return environment.enclosing.Get(name)
	}
	return nil, undefinedVariableError(name)
}

func undefinedVariableError(name *Token) error {
	return RuntimeError{
		Token:   name,
		Message: "Undefined variable '" + name.Lexeme + "'.",
	}
}
