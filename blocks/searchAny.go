package blocks

import (
	"errors"
	"reflect"
	"strconv"
)

type JSONType interface {
	int | ~string | bool
}

func searchArray(search string, input []any, nextKeys []string) (retVal any, err error) {
	index, err := strconv.Atoi(search)
	if err != nil {
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
				Logger.Sugar().Warn("failure fetching value in array", "array", vSlice)
				return vSlice, err
			}
			Logger.Sugar().Debugw("return value", "res", retVal)

		default:
			Logger.Sugar().Debugw("value is not an expected type, continuing on", "item", item, "type", valRef.Kind())
		}
	}
	return retVal, errors.New("failed to find key in provided map")
}

// Find output in the interface we consume from exec output
// takes as input the key as string to gather
// TODO: extend this out to handle arbitrary arrays and maps in arrays
// Returns a json of
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
