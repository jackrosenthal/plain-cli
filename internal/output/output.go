package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

type Printer interface {
	Print(v any) error
	PrintList(v any) error
}

func New(format Format, w io.Writer) Printer {
	switch format {
	case FormatJSON:
		return &jsonPrinter{w: w}
	case FormatPlain:
		return &plainPrinter{w: w}
	case FormatTable:
		fallthrough
	default:
		return &tablePrinter{w: w}
	}
}

type fieldValue struct {
	Key   string
	Label string
	Value string
}

func recordFields(v any) ([]fieldValue, error) {
	value := indirectValue(reflect.ValueOf(v))
	if !value.IsValid() {
		return nil, nil
	}

	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("output: expected struct, got %s", value.Kind())
	}

	typ := value.Type()
	fields := make([]fieldValue, 0, value.NumField())
	for idx := range value.NumField() {
		structField := typ.Field(idx)
		if structField.PkgPath != "" {
			continue
		}

		key, ok := fieldKey(structField)
		if !ok {
			continue
		}

		fields = append(fields, fieldValue{
			Key:   key,
			Label: fieldLabel(structField.Name),
			Value: stringifyValue(value.Field(idx)),
		})
	}

	return fields, nil
}

func listRecords(v any) ([][]fieldValue, error) {
	value := indirectValue(reflect.ValueOf(v))
	if !value.IsValid() {
		return nil, nil
	}

	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return nil, fmt.Errorf("output: expected slice or array, got %s", value.Kind())
	}

	records := make([][]fieldValue, 0, value.Len())
	for idx := range value.Len() {
		fields, err := recordFields(value.Index(idx).Interface())
		if err != nil {
			return nil, err
		}

		records = append(records, fields)
	}

	return records, nil
}

func fieldKey(structField reflect.StructField) (string, bool) {
	tag := structField.Tag.Get("json")
	if tag == "-" {
		return "", false
	}

	if name := strings.TrimSpace(strings.Split(tag, ",")[0]); name != "" {
		return name, true
	}

	name := structField.Name
	if name == "" {
		return "", false
	}

	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes), true
}

func fieldLabel(name string) string {
	if name == "" {
		return ""
	}

	var builder strings.Builder
	runes := []rune(name)
	for idx, r := range runes {
		if idx > 0 && unicode.IsUpper(r) {
			prev := runes[idx-1]
			nextLower := idx+1 < len(runes) && unicode.IsLower(runes[idx+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				builder.WriteByte(' ')
			}
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

func stringifyValue(value reflect.Value) string {
	value = indirectValue(value)
	if !value.IsValid() {
		return ""
	}

	if raw, ok := value.Interface().(json.RawMessage); ok {
		return string(raw)
	}

	if stringer, ok := value.Interface().(fmt.Stringer); ok {
		return stringer.String()
	}

	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", value.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	case reflect.Slice, reflect.Array:
		if value.Len() == 0 {
			return ""
		}

		if isScalarCollection(value) {
			parts := make([]string, 0, value.Len())
			for idx := range value.Len() {
				parts = append(parts, stringifyValue(value.Index(idx)))
			}

			return strings.Join(parts, "\n")
		}

		return marshalInlineJSON(value.Interface())
	case reflect.Struct, reflect.Map:
		return marshalInlineJSON(value.Interface())
	case reflect.Interface:
		return stringifyValue(value.Elem())
	default:
		return fmt.Sprintf("%v", value.Interface())
	}
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}

		value = value.Elem()
	}

	return value
}

func isScalarCollection(value reflect.Value) bool {
	elem := value.Type().Elem()
	for elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	switch elem.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func marshalInlineJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}

	var out bytes.Buffer
	if err := json.Compact(&out, data); err != nil {
		return string(data)
	}

	return out.String()
}
