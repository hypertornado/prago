package pragoai

import (
	"fmt"
	"reflect"
	"strings"
)

// jsonFieldName returns the JSON property name for a struct field, based on
// its `json` tag, and whether the field is optional (has "omitempty"). Name
// is "" when the field should be skipped (json:"-" or unexported).
func jsonFieldName(field reflect.StructField) (name string, omitempty bool) {
	if field.PkgPath != "" {
		return "", false
	}

	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = field.Name
	}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty
}

// fieldDescription returns the description to use for a struct field in
// the generated schema: the `ai` tag if present (e.g. `ai:"formát
// RRRR-MM-DD"`), otherwise falling back to the `prago-name` tag already
// used elsewhere in the codebase for a human-readable field label.
func fieldDescription(field reflect.StructField) string {
	if description := field.Tag.Get("ai"); description != "" {
		return description
	}
	return field.Tag.Get("prago-name")
}

// structSchema builds a JSON schema describing t, for use as a tool input
// schema. It recurses into pointers, slices/arrays and nested structs.
// Supported field kinds: string, bool, ints, floats, slices/arrays of
// those, and nested structs. A struct field can be described for the model
// with the `ai` tag, e.g. `ai:"formát RRRR-MM-DD"`, falling back to
// `prago-name` when `ai` is not set.
func structSchema(t reflect.Type) (map[string]any, error) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}, nil
	case reflect.Bool:
		return map[string]any{"type": "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}, nil
	case reflect.Slice, reflect.Array:
		items, err := structSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]any{"type": "array", "items": items}, nil
	case reflect.Struct:
		properties := map[string]any{}
		var required []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name, omitempty := jsonFieldName(field)
			if name == "" {
				continue
			}

			fieldSchema, err := structSchema(field.Type)
			if err != nil {
				return nil, fmt.Errorf("pole %s: %w", field.Name, err)
			}
			if description := fieldDescription(field); description != "" {
				fieldSchema["description"] = description
			}

			properties[name] = fieldSchema
			if !omitempty {
				required = append(required, name)
			}
		}
		ret := map[string]any{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			ret["required"] = required
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("nepodporovaný typ %s", t.Kind())
	}
}
