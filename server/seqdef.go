package ubom

import (
	"fmt"
	"strconv"
)

// SeqDefID identifies an immutable sequence-definition version.
type SeqDefID string

type SeqDef struct {
	ID   SeqDefID
	root SeqDefNode
}

// BindableField describes a value supplied to a Bind node in a definition.
// Fields are returned in definition order.
type BindableField struct {
	Name string
	Node SeqDefNode
}

// SeqDefField is the descriptive name for a bindable sequence field.
type SeqDefField = BindableField

func NewSeqDef(root SeqDefNode) SeqDef { return SeqDef{root: root} }

// Validate checks SeqDefNodes for well formed definitions.
func (d SeqDef) Validate() error {
	if d.ID == "" {
		return invalidSeqDef("sequence definition has no ID", nil)
	}
	if err := validateSeqDefNode(d.root); err != nil {
		return err
	}
	return nil
}

func validateSeqDefNode(node SeqDefNode) error {
	if node == nil {
		return invalidSeqDef("sequence definition has no root", nil)
	}

	validateChildren := func(children []SeqDefNode) error {
		if len(children) == 0 {
			return invalidSeqDef("sequence node has no children", nil)
		}
		for _, child := range children {
			if err := validateSeqDefNode(child); err != nil {
				return err
			}
		}
		return nil
	}

	switch n := node.(type) {
	case LiteralNode:
		// TODO check to see if the literal node contains nothing? Do we care about spaces?
		return nil
	case ConcatNode:
		return validateChildren(n.Children)
	case ChoiceNode:
		return validateChildren(n.Children)
	case RangeRadixNode:
		return validateChildren(n.Children)
	case BindNode:
		if err := (FieldName(n.Name)).Validate(); err != nil {
			return invalidSeqDef("bind node has invalid name", err)
		}
		return validateSeqDefNode(n.Child)
	case PlaceSequenceNode:
		if len(n.Alphabets) == 0 {
			return invalidSeqDef("place has no alphabets", nil)
		}
		for _, alphabet := range n.Alphabets {
			if alphabet == "" {
				return invalidSeqDef("place has an empty alphabet", nil)
			}
		}
		return nil
	case RangeNode:
		if err := validateRange(n); err != nil {
			return err
		}
		return nil
	default:
		return invalidSeqDef(fmt.Sprintf("unknown sequence node type %T", node), nil)
	}
}

func validateRange(n RangeNode) error {
	if n.Min < 0 || n.Max < n.Min || n.WidthValue < 1 {
		return invalidSeqDef("invalid range definition", nil)
	}
	if len(strconv.Itoa(n.Max)) > n.WidthValue {
		return invalidSeqDef("invalid range definition", nil)
	}
	if n.Pad != 0 && n.Pad >= '1' && n.Pad <= '9' {
		return invalidSeqDef("invalid range padding character", nil)
	}
	return nil
}

func invalidSeqDef(message string, cause error) error {
	return WrapDomainError(ErrorCategoryDefinition, message, cause)
}

func (d SeqDef) Root() SeqDefNode { return d.root }

// BindableFields returns the fields which may be supplied to Generate.
// A field is listed once, even if a definition contains the same binding name
// in more than one alternative.
func (d SeqDef) BindableFields() []BindableField {
	fields := make([]BindableField, 0)
	seen := map[string]bool{}
	var visit func(SeqDefNode)
	visit = func(node SeqDefNode) {
		switch n := node.(type) {
		case BindNode:
			if !seen[n.Name] {
				seen[n.Name] = true
				fields = append(fields, BindableField{Name: n.Name, Node: n.Child})
			}
			visit(n.Child)
		case ConcatNode:
			for _, child := range n.Children {
				visit(child)
			}
		case ChoiceNode:
			for _, child := range n.Children {
				visit(child)
			}
		case RangeRadixNode:
			for _, child := range n.Children {
				visit(child)
			}
		}
	}
	visit(d.root)
	return fields
}

// Fields is an alias for BindableFields.
func (d SeqDef) Fields() []BindableField { return d.BindableFields() }

// Generate renders the canonical value represented by bindings.
func (d SeqDef) Generate(bindings map[string]string) (string, error) {
	if d.root == nil {
		return "", fmt.Errorf("sequence definition has no root")
	}
	known := map[string]bool{}
	for _, field := range d.BindableFields() {
		known[field.Name] = true
	}
	for name := range bindings {
		if !known[name] {
			return "", fmt.Errorf("unknown binding %q", name)
		}
	}
	value, err := generateSeqDefNode(d.root, bindings)
	if err != nil {
		return "", err
	}
	return value, nil
}

func generateSeqDefNode(node SeqDefNode, bindings map[string]string) (string, error) {
	switch n := node.(type) {
	case LiteralNode:
		return n.Text, nil
	case ConcatNode:
		return generateConcat(n.Children, bindings)
	case RangeRadixNode:
		return generateConcat(n.Children, bindings)
	case ChoiceNode:
		return generateAlternatives(n.Children, bindings)
	case BindNode:
		value, ok := bindings[n.Name]
		if !ok {
			return "", fmt.Errorf("missing binding %q", n.Name)
		}
		trial := cloneBindings(bindings)
		_, next, err := n.Child.parseAt(value, 0, trial)
		if err != nil || next != len(value) {
			if err == nil {
				err = expected(next, "end of binding")
			}
			return "", fmt.Errorf("invalid binding %q: %w", n.Name, err)
		}
		// Parsing the child supplies bindings nested inside this binding. Run
		// generation with those values so nested fields are not silently omitted.
		if _, err := generateSeqDefNode(n.Child, trial); err != nil {
			return "", fmt.Errorf("invalid binding %q: %w", n.Name, err)
		}
		return value, nil
	case PlaceSequenceNode:
		result := make([]byte, len(n.Alphabets))
		for i, alphabet := range n.Alphabets {
			if alphabet == "" {
				return "", fmt.Errorf("place has an empty alphabet")
			}
			result[i] = alphabet[0]
		}
		return string(result), nil
	case RangeNode:
		return n.Format(n.Min)
	default:
		return "", fmt.Errorf("unknown sequence node type %T", node)
	}
}

func generateConcat(nodes []SeqDefNode, bindings map[string]string) (string, error) {
	result := ""
	for _, child := range nodes {
		value, err := generateSeqDefNode(child, bindings)
		if err != nil {
			return "", err
		}
		result += value
	}
	return result, nil
}

func generateAlternatives(nodes []SeqDefNode, bindings map[string]string) (string, error) {
	var lastErr error
	for _, child := range nodes {
		value, err := generateSeqDefNode(child, bindings)
		if err == nil {
			return value, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return "", fmt.Errorf("no sequence alternative can be generated: %w", lastErr)
	}
	return "", fmt.Errorf("sequence has no alternatives")
}

func (d SeqDef) WithID(id SeqDefID) SeqDef {
	d.ID = id
	return d
}

func (d SeqDef) Parse(input string) (string, error) {
	result, err := d.ParseValues(input)
	return result.Value, err
}

// CapturedField is one named segment matched by a Bind node. Source is the
// exact input text. Normalized is an int for a range capture and the source
// string when no more specific structural value is available.
type CapturedField struct {
	Name       string
	Source     string
	Normalized any
}

type ParseResult struct {
	Value  string
	Fields []CapturedField
	// Bindings is a compatibility adapter for callers predating structured
	// captures. It is derived from Fields; duplicate names are rejected.
	Bindings map[string]string
}

type captureState struct {
	fields []CapturedField
	names  map[string]bool
}

func newCaptureState() *captureState { return &captureState{names: map[string]bool{}} }

func (s *captureState) clone() *captureState {
	copyState := newCaptureState()
	copyState.fields = append(copyState.fields, s.fields...)
	for name := range s.names {
		copyState.names[name] = true
	}
	return copyState
}

func (s *captureState) commit(from *captureState) {
	s.fields = append(s.fields[:0], from.fields...)
	s.names = from.names
}

func (d SeqDef) ParseValues(input string) (ParseResult, error) {
	if d.root == nil {
		return ParseResult{}, WrapDomainError(ErrorCategorySyntax, "sequence definition has no root", nil)
	}
	state := newCaptureState()
	value, next, err := parseCapturedNode(d.root, input, 0, state)
	if err != nil {
		return ParseResult{}, WrapDomainError(ErrorCategorySyntax, "sequence input is invalid", err)
	}
	if next != len(input) {
		return ParseResult{}, WrapDomainError(ErrorCategorySyntax, "sequence input is invalid", expected(next, "end of input"))
	}
	bindings := make(map[string]string, len(state.fields))
	for _, field := range state.fields {
		bindings[field.Name] = field.Source
	}
	return ParseResult{Value: value, Fields: state.fields, Bindings: bindings}, nil
}

// parseCapturedNode keeps the existing node interface intact for package-local
// structural nodes, while built-in nodes collect transactional captures.
func parseCapturedNode(node SeqDefNode, input string, offset int, state *captureState) (string, int, error) {
	switch n := node.(type) {
	case LiteralNode:
		return n.parseAt(input, offset, nil)
	case ConcatNode:
		return parseCapturedConcat(n.Children, input, offset, state)
	case RangeRadixNode:
		return parseCapturedConcat(n.Children, input, offset, state)
	case ChoiceNode:
		var lastErr error
		for _, child := range n.Children {
			trial := state.clone()
			value, next, err := parseCapturedNode(child, input, offset, trial)
			if err == nil {
				state.commit(trial)
				return value, next, nil
			}
			lastErr = err
		}
		if lastErr == nil {
			lastErr = expected(offset, "one of the choices")
		}
		return "", offset, lastErr
	case BindNode:
		if state.names[n.Name] {
			return "", offset, fmt.Errorf("duplicate binding %q", n.Name)
		}
		start := offset
		state.names[n.Name] = true
		state.fields = append(state.fields, CapturedField{Name: n.Name})
		index := len(state.fields) - 1
		value, next, err := parseCapturedNode(n.Child, input, offset, state)
		if err != nil {
			return "", start, err
		}
		text := input[start:next]
		state.fields[index].Source = text
		state.fields[index].Normalized = normalizedCapture(n.Child, text)
		return value, next, nil
	case PlaceSequenceNode, RangeNode:
		return node.parseAt(input, offset, nil)
	default:
		bindings := make(map[string]string, len(state.fields))
		for _, field := range state.fields {
			bindings[field.Name] = field.Source
		}
		return node.parseAt(input, offset, bindings)
	}
}

func parseCapturedConcat(nodes []SeqDefNode, input string, offset int, state *captureState) (string, int, error) {
	value := ""
	for _, child := range nodes {
		part, next, err := parseCapturedNode(child, input, offset, state)
		if err != nil {
			return "", offset, err
		}
		value += part
		offset = next
	}
	return value, offset, nil
}

func normalizedCapture(node SeqDefNode, text string) any {
	if rangeNode, ok := node.(RangeNode); ok {
		if value, err := rangeNode.ParseValue(text); err == nil {
			return value
		}
	}
	return text
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
	if err := validateRange(n); err != nil {
		return "", offset, err
	}
	if len(input)-offset < n.WidthValue {
		return "", offset, expected(offset, fmt.Sprintf("%d digits", n.WidthValue))
	}
	text := input[offset : offset+n.WidthValue]
	value, err := n.ParseValue(text)
	if err != nil || value < n.Min || value > n.Max {
		return "", offset, expected(offset, fmt.Sprintf("a number from %d to %d", n.Min, n.Max))
	}
	return text, offset + n.WidthValue, nil
}

func (n RangeNode) Format(value int) (string, error) {
	if err := validateRange(n); err != nil {
		return "", err
	}
	if value < n.Min || value > n.Max {
		return "", fmt.Errorf("value %d is outside range %d-%d", value, n.Min, n.Max)
	}
	digits := strconv.Itoa(value)
	if len(digits) > n.WidthValue {
		return "", fmt.Errorf("invalid range definition: value %d exceeds width %d", value, n.WidthValue)
	}
	pad := n.Pad
	if pad == 0 {
		pad = '0'
	}
	result := make([]byte, n.WidthValue)
	for i := range result {
		result[i] = pad
	}
	copy(result[n.WidthValue-len(digits):], digits)
	return string(result), nil
}

func (n RangeNode) ParseValue(text string) (int, error) {
	if err := validateRange(n); err != nil {
		return 0, err
	}
	if len(text) != n.WidthValue {
		return 0, fmt.Errorf("range value has width %d, want %d", len(text), n.WidthValue)
	}
	pad := n.Pad
	if pad == 0 {
		pad = '0'
	}
	firstDigit := 0
	for firstDigit < len(text) && text[firstDigit] == pad {
		firstDigit++
	}
	if firstDigit == len(text) {
		return 0, nil
	}
	for i := firstDigit; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return 0, fmt.Errorf("range value contains a non-decimal digit")
		}
	}
	return strconv.Atoi(text[firstDigit:])
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
