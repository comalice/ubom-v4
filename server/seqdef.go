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

func (d SeqDef) Root() SeqDefNode { return d.root }

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

type LiteralNode struct{ Text string }

func (n LiteralNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
	if len(input)-offset < len(n.Text) || input[offset:offset+len(n.Text)] != n.Text {
		return "", offset, expected(offset, fmt.Sprintf("%q", n.Text))
	}
	return n.Text, offset + len(n.Text), nil
}

func Literal(text string) SeqDefNode { return LiteralNode{Text: text} }

type ConcatNode struct{ Children []SeqDefNode }

func (n ConcatNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	value := ""
	for _, child := range n.Children {
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
	return ConcatNode{Children: append([]SeqDefNode(nil), nodes...)}
}

type ChoiceNode struct{ Children []SeqDefNode }

func (n ChoiceNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	return parseAlternatives(n.Children, input, offset, bindings)
}

func Choice(nodes ...SeqDefNode) SeqDefNode {
	return ChoiceNode{Children: append([]SeqDefNode(nil), nodes...)}
}

type BranchNode struct{ Children []SeqDefNode }

func (n BranchNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	return parseAlternatives(n.Children, input, offset, bindings)
}

// Branch chooses between complete grammar shapes.
func Branch(nodes ...SeqDefNode) SeqDefNode {
	return BranchNode{Children: append([]SeqDefNode(nil), nodes...)}
}

func parseAlternatives(nodes []SeqDefNode, input string, offset int, bindings map[string]string) (string, int, error) {
	var err error
	for _, child := range nodes {
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

type BindNode struct {
	Name  string
	Child SeqDefNode
}

func (n BindNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	value, next, err := n.Child.parseAt(input, offset, bindings)
	if err != nil {
		return "", offset, err
	}
	bindings[n.Name] = value
	return value, next, nil
}

// Bind records the text matched by child under name.
func Bind(name string, child SeqDefNode) SeqDefNode {
	return BindNode{Name: name, Child: child}
}

type RangeRadixNode struct{ Children []SeqDefNode }

func (n RangeRadixNode) parseAt(input string, offset int, bindings map[string]string) (string, int, error) {
	return ConcatNode{Children: n.Children}.parseAt(input, offset, bindings)
}

// RangeRadix is a sequence of ordered fields.
func RangeRadix(nodes ...SeqDefNode) SeqDefNode {
	return RangeRadixNode{Children: append([]SeqDefNode(nil), nodes...)}
}

type PlaceSequenceNode struct{ Alphabets []string }

func (n PlaceSequenceNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
	start := offset
	for _, alphabet := range n.Alphabets {
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
	return PlaceSequenceNode{Alphabets: append([]string(nil), alphabets...)}
}

type RangeNode struct {
	Min        int
	Max        int
	WidthValue int
	Pad        byte
}

func (n RangeNode) parseAt(input string, offset int, _ map[string]string) (string, int, error) {
	if n.Min < 0 || n.Max < n.Min || n.WidthValue < 1 {
		return "", offset, fmt.Errorf("invalid range definition")
	}
	if len(input)-offset < n.WidthValue {
		return "", offset, expected(offset, fmt.Sprintf("%d digits", n.WidthValue))
	}
	text := input[offset : offset+n.WidthValue]
	for i := range text {
		if text[i] < '0' || text[i] > '9' {
			return "", offset, expected(offset+i, "a decimal digit")
		}
	}
	value, err := strconv.Atoi(text)
	if err != nil || value < n.Min || value > n.Max {
		return "", offset, expected(offset, fmt.Sprintf("a number from %d to %d", n.Min, n.Max))
	}
	return text, offset + n.WidthValue, nil
}

// Range matches a zero-padded decimal number. Its default width is the number
// of digits in max; Width can override it.
func Range(min, max int) RangeNode {
	return RangeNode{Min: min, Max: max, WidthValue: len(strconv.Itoa(max)), Pad: '0'}
}

// Width sets the fixed width and optional padding character for a range.
func (n RangeNode) Width(width int, pad ...byte) RangeNode {
	n.WidthValue = width
	if len(pad) > 0 {
		n.Pad = pad[0]
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
