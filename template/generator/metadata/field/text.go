package field

import "github.com/ironman-project/ironman/text/yaml"

//Text represents a text field
type Text struct {
	Field
}

//NewText returns a new text field
func NewText(field Field) *Text {
	return &Text{field}
}

func (t *Text) String() string {
	return yaml.Print(t)
}