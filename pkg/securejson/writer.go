// pkg/securejson/writer.go
package securejson

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
)

// JSON writes v as JSON with encryption applied to fields tagged secure:"encrypt_id".
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	marshallable, err := toJSONValue(reflect.ValueOf(v))
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  http.StatusInternalServerError,
			"message": "Failed to serialize response",
		})
		return
	}
	_ = json.NewEncoder(w).Encode(marshallable)
}

func toJSONValue(rv reflect.Value) (any, error) {
	if !rv.IsValid() {
		return nil, nil
	}

	// Unwrap pointers/interfaces
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	// IMPORTANT: If the concrete type implements json.Marshaler or encoding.TextMarshaler,
	// delegate to the stdlib so custom types (e.g., custom dates) serialize correctly.
	if rv.CanInterface() {
		v := rv.Interface()

		if _, ok := v.(json.Marshaler); ok {
			return v, nil
		}
		if _, ok := v.(encoding.TextMarshaler); ok {
			return v, nil
		}
		// Also check pointer receiver methods (common for MarshalJSON on custom types)
		if rv.CanAddr() {
			pv := rv.Addr().Interface()
			if _, ok := pv.(json.Marshaler); ok {
				return v, nil
			}
			if _, ok := pv.(encoding.TextMarshaler); ok {
				return v, nil
			}
		}
	}

	switch rv.Kind() {
	case reflect.Struct:
		// Let stdlib handle time.Time (has MarshalJSON)
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			return rv.Interface(), nil
		}

		out := make(map[string]any)
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			ft := rt.Field(i)
			// Skip unexported
			if ft.PkgPath != "" {
				continue
			}
			jsonName, omitEmpty, skip := parseJSONTag(ft.Tag.Get("json"), ft.Name)
			if skip {
				continue
			}

			secureTag := ft.Tag.Get("secure")
			fv := rv.Field(i)

			// Flatten embedded anonymous structs with no explicit json name
			if ft.Anonymous && (jsonName == "" || jsonName == ft.Name) {
				embedded, err := toJSONValue(fv)
				if err != nil {
					return nil, err
				}
				if m, ok := embedded.(map[string]any); ok {
					for k, v := range m {
						out[k] = v
					}
				}
				continue
			}

			if omitEmpty && isZeroValue(fv) {
				continue
			}

			val, err := fieldToJSONValue(fv, secureTag)
			if err != nil {
				return nil, err
			}
			out[jsonName] = val
		}
		return out, nil

	case reflect.Slice, reflect.Array:
		n := rv.Len()
		out := make([]any, n)
		for i := 0; i < n; i++ {
			val, err := toJSONValue(rv.Index(i))
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil

	case reflect.Map:
		if rv.IsNil() {
			return nil, nil
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			// JSON objects require string keys; for others, use fmt-like string
			key := ""
			if k.Kind() == reflect.String {
				key = k.String()
			} else {
				// avoid adding quotes from json.Marshal on keys
				key = strings.Trim(string([]byte(strconv.Quote(fmtAny(k.Interface())))), `"`)
			}

			val, err := toJSONValue(v)
			if err != nil {
				return nil, err
			}
			out[key] = val
		}
		return out, nil

	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.String:
		return rv.String(), nil
	default:
		// Let json handle other kinds (e.g., custom types)
		if rv.CanInterface() {
			return rv.Interface(), nil
		}
		return nil, nil
	}
}

func fieldToJSONValue(fv reflect.Value, secureTag string) (any, error) {
	// Unwrap pointers/interfaces
	for fv.Kind() == reflect.Pointer || fv.Kind() == reflect.Interface {
		if fv.IsNil() {
			return nil, nil
		}
		fv = fv.Elem()
	}

	// Encrypt if tagged
	if secureTag == "encrypt_id" {
		switch fv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s := strconv.FormatInt(fv.Int(), 10)
			enc, err := secureid.EncryptID(s)
			if err != nil {
				return nil, errors.New("failed to encrypt id")
			}
			return enc, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			s := strconv.FormatUint(fv.Uint(), 10)
			enc, err := secureid.EncryptID(s)
			if err != nil {
				return nil, errors.New("failed to encrypt id")
			}
			return enc, nil
		case reflect.String:
			enc, err := secureid.EncryptID(fv.String())
			if err != nil {
				return nil, errors.New("failed to encrypt id")
			}
			return enc, nil
		}
	}

	// Recurse normally
	return toJSONValue(fv)
}

func parseJSONTag(tag, fallback string) (name string, omitEmpty bool, skip bool) {
	if tag == "" {
		return fallback, false, false
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "-" {
		return "", false, true
	}
	name = parts[0]
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitEmpty = true
		}
	}
	if name == "" {
		name = fallback
	}
	return name, omitEmpty, false
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		// Special-case time.Time zero
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		// Generic struct zero check
		z := reflect.Zero(v.Type())
		return reflect.DeepEqual(v.Interface(), z.Interface())
	}
	return false
}

// fmtAny avoids importing fmt just for Sprintf here; inline minimal stringer.
func fmtAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}
