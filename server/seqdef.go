package ubom

import (
	"fmt"
	"strconv"
)

type SeqDefID string

type SeqDef struct {
	ID   SeqDefID
	root SeqDefNode
}

func NewSeqDef(root SeqDefNode) SeqDef { return SeqDef{root: root} }

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
	if d.root == nil {
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

// SeqDefNode is a type-safe grammar node. Constructors below build the AST.
type SeqDefNode interface {
	parseAt(input string, offset int, bindings map[string]string) (string, int, error)
}

type literalNode struct{ text string }

func (n literalNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
	if len(input)-offset < len(n.text) || input[offset:offset+len(n.text)] != n.text {
		return "", offset, expected(offset, fmt.Sprintf("%q", n.text))
	}
	return n.text, offset + len(n.text), nil
}

func Literal(text string) SeqDefNode { return literalNode{text: text} }

type concatNode struct{ children []SeqDefNode }

func (n concatNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
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
}

func Concat(nodes ...SeqDefNode) SeqDefNode {
	return concatNode{children: append([]SeqDefNode(nil), nodes...)}
}

type choiceNode struct{ children []SeqDefNode }

func (n choiceNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
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
}

// Choice chooses between parser alternatives.
func Choice(nodes ...SeqDefNode) SeqDefNode {
	return choiceNode{children: append([]SeqDefNode(nil), nodes...)}
}

type branchNode struct{ children []SeqDefNode }

func (n branchNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	return choiceNode{children: n.children}.parseAt(input, offset, bindings)
}

// Branch chooses between complete grammar shapes.
func Branch(nodes ...SeqDefNode) SeqDefNode {
	return branchNode{children: append([]SeqDefNode(nil), nodes...)}
}

type bindNode struct {
	name  string
	child SeqDefNode
}

func (n bindNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	value, next, err := n.child.parseAt(input, offset, bindings)
	if err != nil {
		return "", offset, err
	}
	bindings[n.name] = value
	return value, next, nil
}

// Bind records the text matched by child under name.
func Bind(name string, child SeqDefNode) SeqDefNode {
	return bindNode{name: name, child: child}
}

type rangeRadixNode struct{ children []SeqDefNode }

func (n rangeRadixNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	return concatNode{children: n.children}.parseAt(input, offset, bindings)
}

// RangeRadix is a sequence of ordered fields.
func RangeRadix(nodes ...SeqDefNode) SeqDefNode {
	return rangeRadixNode{children: append([]SeqDefNode(nil), nodes...)}
}

type placeSequenceNode struct{ alphabets []string }

func (n placeSequenceNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
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
}

func PlaceSequence(alphabets ...string) SeqDefNode {
	return placeSequenceNode{alphabets: append([]string(nil), alphabets...)}
}

type rangeNode struct {
	min   int
	max   int
	width int
	pad   byte
}

func (n rangeNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
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

// Range matches a zero-padded decimal number. Its default width is the number
// of digits in max; Width can override it.
func Range(min, max int) rangeNode {
	return rangeNode{min: min, max: max, width: len(strconv.Itoa(max)), pad: '0'}
}

// Width sets the fixed width and optional padding character for a range.
func (n rangeNode) Width(width int, pad ...byte) rangeNode {
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
