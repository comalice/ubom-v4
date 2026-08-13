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
	if d.root == nil {
		return "", fmt.Errorf("sequence definition has no root")
	}
	value, next, err := d.root(input, 0)
	if err != nil {
		return "", err
	}
	if next != len(input) {
		return "", expected(next, "end of input")
	}
	return value, nil
}

// SeqNode parses input at offset and returns the matched text and next offset.
type SeqNode func(input string, offset int) (value string, next int, err error)

func Literal(text string) SeqNode {
	return func(input string, offset int) (string, int, error) {
		if len(input)-offset < len(text) || input[offset:offset+len(text)] != text {
			return "", offset, expected(offset, fmt.Sprintf("%q", text))
		}
		return text, offset + len(text), nil
	}
}

func Concat(nodes ...SeqNode) SeqNode {
	return func(input string, offset int) (string, int, error) {
		value := ""
		for _, node := range nodes {
			part, next, err := node(input, offset)
			if err != nil {
				return "", offset, err
			}
			value += part
			offset = next
		}
		return value, offset, nil
	}
}

// Choice tries alternatives from left to right.
func Choice(nodes ...SeqNode) SeqNode {
	return func(input string, offset int) (string, int, error) {
		var err error
		for _, node := range nodes {
			value, next, nodeErr := node(input, offset)
			if nodeErr == nil {
				return value, next, nil
			}
			err = nodeErr
		}
		if err == nil {
			err = expected(offset, "one of the choices")
		}
		return "", offset, err
	}
}

// PlaceSequence consumes one character from each alphabet. Alphabet lengths
// are the radices of the places.
func PlaceSequence(alphabets ...string) SeqNode {
	return func(input string, offset int) (string, int, error) {
		start := offset
		for _, alphabet := range alphabets {
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
