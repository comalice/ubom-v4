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
	result, err := d.ParseValues(input)
	return result.Value, err
}

type ParseResult struct {
	Value    string
	Bindings map[string]string
}

func (d SeqDef) ParseValues(input string) (ParseResult, error) {
	if d.root.kind == nodeInvalid {
		return ParseResult{}, fmt.Errorf("sequence definition has no root")
	}

	bindings := map[string]string{}
	value, next, err := d.root.parseAt(input, 0, bindings)
	if err != nil {
		return ParseResult{}, err
	}
	if next != len(input) {
		return ParseResult{}, expected(next, "end of input")
	}
	return ParseResult{Value: value, Bindings: bindings}, nil
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
	nodeBranch
	nodeBind
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
	name      string
	child     *SeqNode
}

func (n SeqNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	switch n.kind {
	case nodeLiteral:
		if len(input)-offset < len(n.text) || input[offset:offset+len(n.text)] != n.text {
			return "", offset, expected(offset, fmt.Sprintf("%q", n.text))
		}
		return n.text, offset + len(n.text), nil

	case nodeConcat, nodeRangeRadix:
		value := ""
		for _, child := range n.children {
			part, next, err := child.parseAt(input, offset, bindings)
			if err != nil {
				return "", offset, err
			}
			value += part
			offset = next
		}
		return value, offset, nil

	case nodeChoice, nodeBranch:
		var err error
		for _, child := range n.children {
			trial := cloneBindings(bindings)
			value, next, childErr := child.parseAt(input, offset, trial)
			if childErr == nil {
				copyBindings(bindings, trial)
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

	case nodeBind:
		value, next, err := n.child.parseAt(input, offset, bindings)
		if err != nil {
			return "", offset, err
		}
		bindings[n.name] = value
		return value, next, nil

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

// Branch chooses between complete grammar shapes. It currently parses like
// Choice, but keeps a distinct AST kind for future branch-aware generation.
func Branch(nodes ...SeqNode) SeqNode {
	return SeqNode{kind: nodeBranch, children: append([]SeqNode(nil), nodes...)}
}

// Bind records the text matched by child under name.
func Bind(name string, child SeqNode) SeqNode {
	return SeqNode{kind: nodeBind, name: name, child: &child}
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

func cloneBindings(bindings map[string]string) map[string]string {
	clone := make(map[string]string, len(bindings))
	copyBindings(clone, bindings)
	return clone
}

func copyBindings(dst, src map[string]string) {
	for name, value := range src {
		dst[name] = value
	}
}
