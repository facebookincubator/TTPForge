package atomics

import (
	"errors"
	"reflect"
	"strconv"

	"go.uber.org/zap"
)

// JSONType is a custom interface that represents a JSON value, which can be one of the following types:
// integer, string, or boolean. This interface is used to define a variable or a struct field that can
// hold any of these three types, allowing more flexibility when working with JSON data.
//
// Example usage:
//
// type ExampleStruct struct {
// Field JSONType
// }
//
// var exampleVar JSONType = 42
// exampleVar = "Hello, World!"
// exampleVar = true
type JSONType interface {
	int | ~string | bool
}

func searchArray(search string, input []any, nextKeys []string) (retVal any, err error) {
	index, err := strconv.Atoi(search)
	if err != nil {
		Logger.Error("failed to convert string to int", zap.Error(err))
		return retVal, err
	}

	if index > 0 && index < len(input) {
		item := input[index]

		// Check if value in search is the last one and no nextKeys are left to find
		switch valRef := reflect.ValueOf(item); valRef.Kind() {
		case reflect.Bool:
			Logger.Sugar().Debugw("value for key is bool", "val", valRef.Bool())
			return valRef.Bool(), nil
		case reflect.String:
			Logger.Sugar().Debugw("value for key is string", "val", valRef.String())
			return valRef.String(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			Logger.Sugar().Debugw("value for key is int", "val", valRef.Int())
			return valRef.Int(), nil
		case reflect.Map:
			Logger.Sugar().Debugw("value for key is map", "item", item)
			retVal, err = mapCase(nextKeys, item)
			if err != nil {
				Logger.Sugar().Warnw("issue unpacking nested map, returning early", "item", item, "index", search)
				return
			}
		case reflect.Slice:
			Logger.Sugar().Debugw("value for index is array", "val", item)
			vSlice, ok := item.([]any)
			if !ok {
				Logger.Sugar().Warnw("issue indexing into array, returning early", "item", item, "index", search)
				return item, errors.New("failed to unpack nested map")
			}

			Logger.Sugar().Debugw("value for index is array", "val", item)
			retVal, err = searchArray(nextKeys[0], vSlice, nextKeys[1:])
			if err != nil {
				Logger.Sugar().Errorw("failure fetching value in array", "array", vSlice)
				return vSlice, err
			}
			Logger.Sugar().Debugw("return value", "res", retVal)

		default:
			Logger.Sugar().Debugw("value is not an expected type, continuing on", "item", item, "type", valRef.Kind())
		}
	}
	Logger.Error("failed to find key in provided map", zap.Error(err))
	return retVal, errors.New("failed to find key in provided map")
}

// TODO: extend this out to handle arbitrary arrays and maps in arrays
// searchMap recursively searches for a key within a nested JSON object and returns the associated value.
// The function accepts a search key, an input map, and a slice of nextKeys representing the remaining
// keys to search for in nested maps or arrays. It returns the value associated with the search key or an
// error if the key is not found.
//
// The function supports searching within maps and arrays, but may require further extension to handle
// more complex cases such as arbitrary arrays and maps within arrays.
//
// Parameters:
//
// search: The key to search for in the input map.
// input: A map of JSONType keys and 'any' values representing the JSON object to search within.
// nextKeys: A slice of strings representing the remaining keys to search for in nested maps or arrays.
//
// Returns:
//
// any: The value associated with the search key or nil if the key is not found.
// error: An error if the search key is not found, or if there is an issue processing the JSON object.
//
// Example output:
//
// { search : {key : {with : {arbitrary : {depth : val }}}}}
func searchMap[T JSONType](search string, input map[T]any, nextKeys []string) (retVal any, err error) {
	for k, v := range input {
		switch keyRef := reflect.ValueOf(k); keyRef.Kind() {
		case reflect.String:
			if keyRef.String() != search {
				// continue to next iteration
				continue
			}
		default:
			Logger.Sugar().Debugw("key value is not a string, continuing on", "key", k)
			// continue to next iteration
			continue
		}
		// Check if value in search is the last one and no nextKeys are left to find
		switch valRef := reflect.ValueOf(v); valRef.Kind() {
		case reflect.Bool:
			Logger.Sugar().Debugw("value for key is bool", "val", valRef.Bool())
			return valRef.Bool(), nil
		case reflect.String:
			Logger.Sugar().Debugw("value for key is string", "val", valRef.String())
			return valRef.String(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			Logger.Sugar().Debugw("value for key is int", "val", valRef.Int())
			return valRef.Int(), nil
		case reflect.Map:
			Logger.Sugar().Debugw("value for key is map", "val", v)
			retVal, err = mapCase(nextKeys, v)
			if err != nil {
				Logger.Sugar().Warnw("issue unpacking nested map, returning early", "val", v, "key", k)
				return
			}
		case reflect.Slice:
			vSlice, ok := v.([]any)
			if !ok {
				Logger.Sugar().Warnw("issue unpacking nested array, returning early", "val", v, "key", k)
				return v, errors.New("failed to unpack nested map")
			}

			Logger.Sugar().Debugw("value for key is array", "val", v)
			retVal, err = searchArray(nextKeys[0], vSlice, nextKeys[1:])
			if err != nil {
				Logger.Sugar().Warn("failure fetching value in array", "array", vSlice)
				return vSlice, err
			}
			Logger.Sugar().Debugw("return value", "res", retVal)

		default:
			Logger.Sugar().Debugw("value is not an expected type, continuing on", "val", v, "type", valRef.Kind())
		}
	}
	return retVal, errors.New("failed to find key in provided map")
}

func mapCase(nextKeys []string, v any) (retVal any, err error) {
	switch len(nextKeys) {
	case 0:
		// bottom out
		return v, nil
	case 1:
		vMap, ok := v.(map[string]any)
		if !ok {
			return v, errors.New("failed to unpack nested map")
		}
		retVal, err = searchMap(nextKeys[0], vMap, nextKeys[1:])
		if err != nil {
			return vMap, err
		}
		return retVal, nil
	default:
		vMap, ok := v.(map[string]any)
		if !ok {
			return v, errors.New("failed to unpack nested map")
		}
		retVal, err = searchMap(nextKeys[0], vMap, nextKeys[1:])
		if err != nil {
			return vMap, err
		}
		return retVal, nil
	}
}
