package ubom

import (
	"fmt"
	"strconv"
)

type SeqDefID string

// SeqDef is a small grammar for valid part numbers.
type SeqDef struct {
	ID   SeqDefID
	root SeqNode
}

func NewSeqDef(root SeqNode) SeqDef { return SeqDef{root: root} }

func (d SeqDef) WithID(id SeqDefID) SeqDef {
	d.ID = id
	return d
}

func (d SeqDef) Parse(input string) (string, error) {
	if d.root.kind == nodeInvalid {
		return "", fmt.Errorf("sequence definition has no root")
	}
	value, next, err := d.root.parseAt(input, 0)
	if err != nil {
		return "", err
	}
	if next != len(input) {
		return "", expected(next, "end of input")
	}
	return value, nil
}

type nodeKind uint8

const (
	nodeInvalid nodeKind = iota
	nodeLiteral
	nodeConcat
	nodeChoice
	nodePlaces
	nodeRange
	nodeValues
	nodeRangeRadix
)

// SeqNode is the AST. Its fields stay private so nodes are built through the
// small constructors below; later generation can walk this same tree.
type SeqNode struct {
	kind      nodeKind
	text      string
	children  []SeqNode
	alphabets []string
	min       int
	max       int
	width     int
	pad       byte
	values    []string
}

func (n SeqNode) parseAt(input string, offset int) (string, int, error) {
	switch n.kind {
	case nodeLiteral:
		if len(input)-offset < len(n.text) || input[offset:offset+len(n.text)] != n.text {
			return "", offset, expected(offset, fmt.Sprintf("%q", n.text))
		}
		return n.text, offset + len(n.text), nil

	case nodeConcat, nodeRangeRadix:
		value := ""
		for _, child := range n.children {
			part, next, err := child.parseAt(input, offset)
			if err != nil {
				return "", offset, err
			}
			value += part
			offset = next
		}
		return value, offset, nil

	case nodeChoice:
		var err error
		for _, child := range n.children {
			value, next, childErr := child.parseAt(input, offset)
			if childErr == nil {
				return value, next, nil
			}
			err = childErr
		}
		if err == nil {
			err = expected(offset, "one of the choices")
		}
		return "", offset, err

	case nodeValues:
		for _, value := range n.values {
			if len(input)-offset >= len(value) && input[offset:offset+len(value)] == value {
				return value, offset + len(value), nil
			}
		}
		return "", offset, expected(offset, "one of the values")

	case nodePlaces:
		start := offset
		for _, alphabet := range n.alphabets {
			if alphabet == "" {
				return "", start, fmt.Errorf("place has an empty alphabet")
			}
			if offset == len(input) || !contains(alphabet, input[offset]) {
				return "", start, expected(offset, fmt.Sprintf("one of %q", alphabet))
			}
			offset++
		}
		return input[start:offset], offset, nil

	case nodeRange:
		if n.min < 0 || n.max < n.min || n.width < 1 {
			return "", offset, fmt.Errorf("invalid range definition")
		}
		if len(input)-offset < n.width {
			return "", offset, expected(offset, fmt.Sprintf("%d digits", n.width))
		}
		text := input[offset : offset+n.width]
		for i := range text {
			if text[i] < '0' || text[i] > '9' {
				return "", offset, expected(offset+i, "a decimal digit")
			}
		}
		value, err := strconv.Atoi(text)
		if err != nil || value < n.min || value > n.max {
			return "", offset, expected(offset, fmt.Sprintf("a number from %d to %d", n.min, n.max))
		}
		return text, offset + n.width, nil
	}

	return "", offset, fmt.Errorf("invalid sequence node")
}

func Literal(text string) SeqNode {
	return SeqNode{kind: nodeLiteral, text: text}
}

func Concat(nodes ...SeqNode) SeqNode {
	return SeqNode{kind: nodeConcat, children: append([]SeqNode(nil), nodes...)}
}

// Choice tries alternatives from left to right.
func Choice(nodes ...SeqNode) SeqNode {
	return SeqNode{kind: nodeChoice, children: append([]SeqNode(nil), nodes...)}
}

// Values is an ordered set of literal values. Unlike Choice, it is intended
// to remain generation-friendly.
func Values(values ...string) SeqNode {
	return SeqNode{kind: nodeValues, values: append([]string(nil), values...)}
}

// RangeRadix is a sequence of ordered fields. It currently parses like
// Concat, but keeps its own AST kind for future sequencing behavior.
func RangeRadix(nodes ...SeqNode) SeqNode {
	return SeqNode{kind: nodeRangeRadix, children: append([]SeqNode(nil), nodes...)}
}

// PlaceSequence consumes one character from each alphabet. Alphabet lengths
// are the radices of the places.
func PlaceSequence(alphabets ...string) SeqNode {
	return SeqNode{kind: nodePlaces, alphabets: append([]string(nil), alphabets...)}
}

// Range matches a zero-padded decimal number. Its default width is the number
// of digits in max; Width can override it.
func Range(min, max int) SeqNode {
	width := len(strconv.Itoa(max))
	return SeqNode{kind: nodeRange, min: min, max: max, width: width, pad: '0'}
}

// Width sets the fixed width and optional padding character for a range.
// Padding is metadata for generation; parsing currently requires decimal
// digits in every position.
func (n SeqNode) Width(width int, pad ...byte) SeqNode {
	if n.kind != nodeRange {
		return n
	}
	n.width = width
	if len(pad) > 0 {
		n.pad = pad[0]
	}
	return n
}

func contains(alphabet string, value byte) bool {
	for i := range alphabet {
		if alphabet[i] == value {
			return true
		}
	}
	return false
}

func expected(position int, what string) error {
	return fmt.Errorf("invalid sequence at position %d: expected %s", position, what)
}
