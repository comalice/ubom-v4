package store

import (
	"encoding/json"
	"fmt"

	ubom "ubom-v4"
)

type seqDefRecord struct {
	Kind      string         `json:"kind"`
	Text      string         `json:"text,omitempty"`
	Name      string         `json:"name,omitempty"`
	Min       int            `json:"min,omitempty"`
	Max       int            `json:"max,omitempty"`
	Width     int            `json:"width,omitempty"`
	Pad       byte           `json:"pad,omitempty"`
	Alphabets []string       `json:"alphabets,omitempty"`
	Children  []seqDefRecord `json:"children,omitempty"`
	Child     *seqDefRecord  `json:"child,omitempty"`
}

func marshalSeqDef(def ubom.SeqDef) ([]byte, error) {
	node, err := marshalSeqDefNode(def.Root())
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		ID   ubom.SeqDefID `json:"id"`
		Root seqDefRecord  `json:"root"`
	}{ID: def.ID, Root: node})
}

func marshalSeqDefNode(node ubom.SeqDefNode) (seqDefRecord, error) {
	switch n := node.(type) {
	case ubom.LiteralNode:
		return seqDefRecord{Kind: "literal", Text: n.Text}, nil
	case ubom.ConcatNode:
		return marshalChildren("concat", n.Children)
	case ubom.ChoiceNode:
		return marshalChildren("choice", n.Children)
	case ubom.BranchNode:
		return marshalChildren("branch", n.Children)
	case ubom.RangeRadixNode:
		return marshalChildren("range_radix", n.Children)
	case ubom.BindNode:
		child, err := marshalSeqDefNode(n.Child)
		if err != nil {
			return seqDefRecord{}, err
		}
		return seqDefRecord{Kind: "bind", Name: n.Name, Child: &child}, nil
	case ubom.PlaceSequenceNode:
		return seqDefRecord{Kind: "places", Alphabets: n.Alphabets}, nil
	case ubom.RangeNode:
		return seqDefRecord{Kind: "range", Min: n.Min, Max: n.Max, Width: n.WidthValue, Pad: n.Pad}, nil
	default:
		return seqDefRecord{}, fmt.Errorf("cannot serialize unknown sequence node %T", node)
	}
}

func marshalChildren(kind string, children []ubom.SeqDefNode) (seqDefRecord, error) {
	record := seqDefRecord{Kind: kind, Children: make([]seqDefRecord, 0, len(children))}
	for _, child := range children {
		encoded, err := marshalSeqDefNode(child)
		if err != nil {
			return seqDefRecord{}, err
		}
		record.Children = append(record.Children, encoded)
	}
	return record, nil
}

func unmarshalSeqDef(data []byte) (ubom.SeqDef, error) {
	var value struct {
		ID   ubom.SeqDefID `json:"id"`
		Root seqDefRecord  `json:"root"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return ubom.SeqDef{}, err
	}
	root, err := unmarshalSeqDefNode(value.Root)
	if err != nil {
		return ubom.SeqDef{}, err
	}
	return ubom.NewSeqDef(root).WithID(value.ID), nil
}

func unmarshalSeqDefNode(record seqDefRecord) (ubom.SeqDefNode, error) {
	children := func() ([]ubom.SeqDefNode, error) {
		result := make([]ubom.SeqDefNode, 0, len(record.Children))
		for _, child := range record.Children {
			node, err := unmarshalSeqDefNode(child)
			if err != nil {
				return nil, err
			}
			result = append(result, node)
		}
		return result, nil
	}

	switch record.Kind {
	case "literal":
		return ubom.LiteralNode{Text: record.Text}, nil
	case "concat":
		value, err := children()
		return ubom.ConcatNode{Children: value}, err
	case "choice":
		value, err := children()
		return ubom.ChoiceNode{Children: value}, err
	case "branch":
		value, err := children()
		return ubom.BranchNode{Children: value}, err
	case "range_radix":
		value, err := children()
		return ubom.RangeRadixNode{Children: value}, err
	case "bind":
		if record.Child == nil {
			return nil, fmt.Errorf("bind node has no child")
		}
		child, err := unmarshalSeqDefNode(*record.Child)
		return ubom.BindNode{Name: record.Name, Child: child}, err
	case "places":
		return ubom.PlaceSequenceNode{Alphabets: record.Alphabets}, nil
	case "range":
		return ubom.RangeNode{Min: record.Min, Max: record.Max, WidthValue: record.Width, Pad: record.Pad}, nil
	default:
		return nil, fmt.Errorf("unknown sequence node kind %q", record.Kind)
	}
}
