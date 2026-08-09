package compute

import (
	"errors"
	"strings"
)

const (
	UnknownCommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

var (
	ErrInvalidCommand   = errors.New("invalid command")
	ErrInvalidArguments = errors.New("invalid arguments")
)

type Query interface {
	CommandID() int
	Arguments() []string
}

type query struct {
	commandID int
	arguments []string
}

func (q query) CommandID() int      { return q.commandID }
func (q query) Arguments() []string { return q.arguments }

type Compute struct{}

func NewCompute() *Compute {
	return &Compute{}
}

func (c *Compute) Parse(queryStr string) (Query, error) {
	fields := strings.Fields(strings.TrimSpace(queryStr))
	if len(fields) == 0 {
		return nil, ErrInvalidCommand
	}

	switch strings.ToUpper(fields[0]) {
	case "SET":
		if len(fields) != 3 {
			return nil, ErrInvalidArguments
		}
		return query{commandID: SetCommandID, arguments: fields[1:]}, nil
	case "GET":
		if len(fields) != 2 {
			return nil, ErrInvalidArguments
		}
		return query{commandID: GetCommandID, arguments: fields[1:]}, nil
	case "DEL":
		if len(fields) != 2 {
			return nil, ErrInvalidArguments
		}
		return query{commandID: DelCommandID, arguments: fields[1:]}, nil
	default:
		return nil, ErrInvalidCommand
	}
}
