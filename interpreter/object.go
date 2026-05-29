package interpreter

import (
	"fmt"
	"strconv"
)

type ObjectType int

const (
	ObjectNil ObjectType = iota
	ObjectNumber
	ObjectString
	ObjectBool
	ObjectFunction
	ObjectArray
)

type Object struct {
	Type  ObjectType
	Value any
}

func NumberObject(v float64) Object {
	return Object{Type: ObjectNumber, Value: v}
}

func NilObject() Object {
	return Object{Type: ObjectNil, Value: nil}
}

func StringObject(v string) Object {
	return Object{Type: ObjectString, Value: v}
}

func BoolObject(v bool) Object {
	return Object{Type: ObjectBool, Value: v}
}

func FunctionObject(v any) Object {
	return Object{Type: ObjectFunction, Value: v}
}

func ArrayObject(v []Object) Object {
	return Object{Type: ObjectArray, Value: v}
}

func (o Object) String() string {
	switch o.Type {
	case ObjectNil:
		return "nil"
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
	case ObjectFunction:
		return "<function>"
	case ObjectArray:
		v, _ := o.Value.([]Object)
		out := "["
		for i, el := range v {
			if i > 0 {
				out += ", "
			}
			out += el.String()
		}
		out += "]"
		return out
	default:
		return fmt.Sprintf("<unknown:%v>", o.Value)
	}
}
