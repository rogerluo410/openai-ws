package server

import (
	"reflect"
	"unicode"
)

// ConvertStructToMap convert the struct to a map which has key with first lowercase letter.
// It skip the zero value of the struct. If you want keep zero value key, put it in the extra map.
func ConvertStructToMap(s reflect.Value, extraMap ...map[string]interface{}) map[string]interface{} {
	var jsonMap map[string]interface{}

	if s.Kind() == reflect.Ptr {
		s = reflect.Indirect(s)
	}

	if s.Kind() != reflect.Struct {
		return nil
	}

	if len(extraMap) == 1 {
		jsonMap = extraMap[0]
	} else {
		jsonMap = make(map[string]interface{})
	}

	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		key := lowercaseFirstLetter(s.Type().Field(i).Name)
		if _, hasKey := jsonMap[key]; hasKey {
			continue
		}

		if field.Type().Kind() == reflect.Ptr && !field.IsNil() {
			field = reflect.Indirect(field)
		}

		switch {
		case isZero(field):
			break
		case field.Type().Kind() == reflect.Struct:
			jsonMap[key] = ConvertStructToMap(field)
		default:
			jsonMap[key] = field.Interface()
		}
	}

	return jsonMap
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		return v.IsNil()
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	case reflect.Int:
		return false
	default:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

func lowercaseFirstLetter(s string) string {
	copyStr := []rune(s)
	copyStr[0] = unicode.ToLower(copyStr[0])
	return string(copyStr)
}