package main

type LoxClass struct {
	name    string
	methods map[string]*LoxFunction
}

func NewLoxClass(name string, methods map[string]*LoxFunction) *LoxClass {
	return &LoxClass{name: name, methods: methods}
}

func (klass *LoxClass) Arity() int {
	initializer := klass.FindMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.Arity()
}

func (klass *LoxClass) Call(interpreter *Interpreter, arguments []any) any {
	instance := NewLoxInstance(klass)
	if initializer := klass.FindMethod("init"); initializer != nil {
		initializer.Bind(instance).Call(interpreter, arguments)
	}
	return instance
}

func (klass *LoxClass) FindMethod(name string) *LoxFunction {
	return klass.methods[name]
}

func (klass *LoxClass) String() string {
	return klass.name
}

type LoxInstance struct {
	klass  *LoxClass
	fields map[string]any
}

func NewLoxInstance(klass *LoxClass) *LoxInstance {
	return &LoxInstance{klass: klass, fields: make(map[string]any)}
}

func (instance *LoxInstance) Get(name *Token) (any, error) {
	if value, ok := instance.fields[name.Lexeme]; ok {
		return value, nil
	}
	if method := instance.klass.FindMethod(name.Lexeme); method != nil {
		return method.Bind(instance), nil
	}
	return nil, RuntimeError{Token: name, Message: "Undefined property '" + name.Lexeme + "'."}
}

func (instance *LoxInstance) Set(name *Token, value any) {
	instance.fields[name.Lexeme] = value
}

func (instance *LoxInstance) String() string {
	return instance.klass.name + " instance"
}
