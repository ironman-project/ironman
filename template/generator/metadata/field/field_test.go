package field

import (
	"reflect"
	"testing"
)

func TestField_ID(t *testing.T) {
	tests := []struct {
		name string
		f    Field
		want string
	}{
		{"Get right ID", Field{"id": "myID"}, "myID"},
		{"Get empty  if field ID is nil", Field{}, ""},
		{"Get empty  if field ID is cannot be asserted to string", Field{"id": 1}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ID(); got != tt.want {
				t.Errorf("%q. Field.ID() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestField_Label(t *testing.T) {
	tests := []struct {
		name string
		f    Field
		want string
	}{
		{"Get right Label", Field{"label": "My label"}, "My label"},
		{"Get empty  if field label is nil", Field{}, ""},
		{"Get empty  if field label is cannot be asserted to string", Field{"label": 1}, ""},
	}
	for _, tt := range tests {
		if got := tt.f.Label(); got != tt.want {
			t.Run(tt.name, func(t *testing.T) {
				t.Errorf("%q. Field.Label() = %v, want %v", tt.name, got, tt.want)
			})

		}
	}
}

func TestField_Type(t *testing.T) {
	tests := []struct {
		name string
		f    Field
		want string
	}{
		{"Get right Label", Field{"type": "thetype"}, "thetype"},
		{"Get empty  if field type is nil", Field{}, ""},
		{"Get empty  if field type is cannot be asserted to string", Field{"type": 1}, ""},
	}
	for _, tt := range tests {
		if got := tt.f.Type(); got != tt.want {
			t.Errorf("%q. Field.Type() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestField_stringValue(t *testing.T) {
	type args struct {
		fieldName string
	}
	tests := []struct {
		name string
		f    Field
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := tt.f.stringValue(tt.args.fieldName); got != tt.want {
			t.Errorf("%q. Field.stringValue() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestValidateMandatoryFieldAttributes(t *testing.T) {
	type args struct {
		f Field
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Validation error id", args{f: Field{}}, true},
		{"Validation error type", args{f: Field{"id": "something"}}, true},
		{"Validation error label", args{f: Field{"id": "something", "type": "sometype"}}, true},
		{"Valid field", args{f: Field{"id": "anId", "type": "thetype", "label": "this is a random label"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateMandatoryFieldAttributes(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("%q. ValidateMandatoryFieldAttributes() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})

	}
}

func TestMapUnstructuredToField(t *testing.T) {
	type args struct {
		unstructuredField interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			"Text field",
			args{
				map[string]interface{}{
					"id":    "myTextField",
					"type":  string(TypeText),
					"label": "My text field",
				},
			},
			NewText(
				Field{
					"id":    "myTextField",
					"type":  string(TypeText),
					"label": "My text field",
				},
			),
			false,
		},
		{
			"Array",
			args{
				map[string]interface{}{
					"id":    "myFixedArray",
					"type":  string(TypeArray),
					"label": "My fixed array",
					"size":  3,
					"field_definition": map[string]interface{}{
						"type":  "text",
						"label": "enter a value",
					},
				}},
			NewArray(
				Field(map[string]interface{}{
					"id":    "myFixedArray",
					"type":  string(TypeArray),
					"label": "My fixed array",
					"size":  3,
					"field_definition": map[string]interface{}{
						"id":    "placeholder",
						"type":  "text",
						"label": "enter a value",
					},
				}),
				3,
				NewText(
					Field{
						"id":    "placeholder",
						"type":  string(TypeText),
						"label": "enter a value",
					},
				),
			),
			false,
		},
		{
			"Fixed list",
			args{
				map[string]interface{}{
					"id":    "myFixedList",
					"type":  string(TypeFixedList),
					"label": "My fixed list",
					"fields": []interface{}{
						map[string]interface{}{
							"id":    "myTextField1",
							"type":  string(TypeText),
							"label": "My text field1",
						},
						map[string]interface{}{
							"id":    "myTextField2",
							"type":  string(TypeText),
							"label": "My text field2",
						},
						map[string]interface{}{
							"id":    "myTextField3",
							"type":  string(TypeText),
							"label": "My text field3",
						},
					},
				},
			},
			NewFixedList(
				Field(
					map[string]interface{}{
						"id":    "myFixedList",
						"type":  string(TypeFixedList),
						"label": "My fixed list",
						"fields": []interface{}{
							map[string]interface{}{
								"id":    "myTextField1",
								"type":  string(TypeText),
								"label": "My text field1",
							},
							map[string]interface{}{
								"id":    "myTextField2",
								"type":  string(TypeText),
								"label": "My text field2",
							},
							map[string]interface{}{
								"id":    "myTextField3",
								"type":  string(TypeText),
								"label": "My text field3",
							},
						},
					}),
				[]interface{}{
					NewText(
						Field{
							"id":    "myTextField1",
							"type":  string(TypeText),
							"label": "My text field1",
						},
					),
					NewText(
						Field{
							"id":    "myTextField2",
							"type":  string(TypeText),
							"label": "My text field2",
						},
					),
					NewText(
						Field{
							"id":    "myTextField3",
							"type":  string(TypeText),
							"label": "My text field3",
						},
					),
				},
			),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapUnstructuredToField(tt.args.unstructuredField)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. MapUnstructuredToField() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. MapUnstructuredToField() = \n%v, \nwant \n%v", tt.name, got, tt.want)
			}
		})

	}
}
