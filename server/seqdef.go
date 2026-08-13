package ubom

import "fmt"

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
)

// SeqNode is the AST. Its fields stay private so nodes are built through the
// small constructors below; later generation can walk this same tree.
type SeqNode struct {
	kind      nodeKind
	text      string
	children  []SeqNode
	alphabets []string
}

func (n SeqNode) parseAt(input string, offset int) (string, int, error) {
	switch n.kind {
	case nodeLiteral:
		if len(input)-offset < len(n.text) || input[offset:offset+len(n.text)] != n.text {
			return "", offset, expected(offset, fmt.Sprintf("%q", n.text))
		}
		return n.text, offset + len(n.text), nil

	case nodeConcat:
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

// PlaceSequence consumes one character from each alphabet. Alphabet lengths
// are the radices of the places.
func PlaceSequence(alphabets ...string) SeqNode {
	return SeqNode{kind: nodePlaces, alphabets: append([]string(nil), alphabets...)}
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
