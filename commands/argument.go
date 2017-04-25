package commands

type ArgumentType int

const (
	ArgString ArgumentType = iota
	ArgFile
)

// Argument comment to be added
type Argument struct {
	Name          string
	Type          ArgumentType
	Required      bool
	Variadic      bool
	SupportsStdin bool
	Recursive     bool
	Description   string
}

func (a Argument) EnableStdin() Argument {
	a.SupportsStdin = true
	return a
}

func StringArg(name string, required, variadic bool, description string) Argument {
	return Argument{
		Name:        name,
		Type:        ArgString,
		Required:    required,
		Variadic:    variadic,
		Description: description,
	}
}

func FileArg(name string, required, variadic bool, description string) Argument {
	return Argument{
		Name:        name,
		Type:        ArgFile,
		Required:    required,
		Variadic:    variadic,
		Description: description,
	}
}
