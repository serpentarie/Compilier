package interpreter

import (
	"fmt"
	"strconv"
)

type ObjectType int

const (
	ObjectNumber ObjectType = iota
	ObjectString
	ObjectBool
)

type Object struct {
	Type  ObjectType
	Value any
}

func NumberObject(v float64) Object {
	return Object{Type: ObjectNumber, Value: v}
}

func StringObject(v string) Object {
	return Object{Type: ObjectString, Value: v}
}

func BoolObject(v bool) Object {
	return Object{Type: ObjectBool, Value: v}
}

func (o Object) String() string {
	switch o.Type {
	case ObjectNumber:
		v, _ := o.Value.(float64)
		return strconv.FormatFloat(v, 'f', -1, 64)
	case ObjectString:
		v, _ := o.Value.(string)
		return v
	case ObjectBool:
		v, _ := o.Value.(bool)
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("<unknown:%v>", o.Value)
	}
}
