package ast

import (
	"bytes"
	"fmt"
)

type List struct {
	Not   bool
	Chars string
}

type POSIX struct {
	Not   bool
	Class string
}

type Range struct {
	Not    bool
	Lo, Hi rune
}

type Text struct {
	Text string
}

type Capture struct {
	Quantifier string
}

type Kind int

const (
	KindNothing Kind = iota
	KindPattern
	KindList
	KindPOSIX
	KindRange
	KindCapture
	KindText
	KindAny
	KindSuper
	KindSingle
	KindAnyOf
)

type Node struct {
	Parent   *Node
	Children []*Node
	Value    interface{}
	Kind     Kind
}

func NewNode(k Kind, v interface{}, ch ...*Node) *Node {
	n := &Node{
		Kind:  k,
		Value: v,
	}
	for _, c := range ch {
		Insert(n, c)
	}
	return n
}

func Insert(parent *Node, children ...*Node) {
	parent.Children = append(parent.Children, children...)
	for _, ch := range children {
		ch.Parent = parent
	}
}

func (a *Node) String() string {
	var buf bytes.Buffer
	buf.WriteString(a.Kind.String())
	if a.Value != nil {
		buf.WriteString(" =")
		buf.WriteString(fmt.Sprintf("%v", a.Value))
	}
	if len(a.Children) > 0 {
		buf.WriteString(" [")
		for i, c := range a.Children {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c.String())
		}
		buf.WriteString("]")
	}
	return buf.String()
}

func (k Kind) String() string {
	switch k {
	case KindNothing:
		return "Nothing"
	case KindPattern:
		return "Pattern"
	case KindList:
		return "List"
	case KindPOSIX:
		return "POSIX"
	case KindRange:
		return "Range"
	case KindCapture:
		return "Capture"
	case KindText:
		return "Text"
	case KindAny:
		return "Any"
	case KindSuper:
		return "Super"
	case KindSingle:
		return "Single"
	case KindAnyOf:
		return "AnyOf"
	default:
		return ""
	}
}
