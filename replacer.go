package scimpatch

import (
	"github.com/scim2/filter-parser/v2"
)

type replacer struct{}

var replacerInstance *replacer

func (r *replacer) Direct(scopedMap map[string]interface{}, scopedAttr string, value interface{}) (map[string]interface{}, bool) {
	switch newValue := value.(type) {
	case []map[string]interface{}:
		oldSlice, ok := scopedMap[scopedAttr]
		if !ok {
			scopedMap[scopedAttr] = newValue
			return scopedMap, true
		}
		oldMaps, ok := areEveryItemsMap(oldSlice)
		if !ok {
			// WARN: unexpected current value
			scopedMap[scopedAttr] = newValue
			return scopedMap, true
		}
		if len(oldMaps) != len(newValue) {
			scopedMap[scopedAttr] = newValue
			return scopedMap, true
		}
		for _, newMap := range newValue {
			found := false
			for _, oldMap := range oldMaps {
				if eqMap(newMap, oldMap) {
					found = true
					break
				}
			}
			if !found {
				scopedMap[scopedAttr] = newValue
				return scopedMap, true
			}
		}
		return scopedMap, false
	case map[string]interface{}:
		oldMap, ok := scopedMap[scopedAttr].(map[string]interface{})
		if ok && eqMap(newValue, oldMap) {
			return scopedMap, false
		}
		scopedMap[scopedAttr] = value
		return scopedMap, true
	case []interface{}:
		oldSlice, ok := scopedMap[scopedAttr].([]interface{})
		if !ok {
			scopedMap[scopedAttr] = newValue
			return scopedMap, true
		}
		if len(oldSlice) != len(newValue) {
			scopedMap[scopedAttr] = newValue
			return scopedMap, true
		}
		if oldMaps, ok := areEveryItemsMap(oldSlice); ok {
			if newMaps, ok := areEveryItemsMap(newValue); ok {
				for _, newMap := range newMaps {
					found := false
					for _, oldMap := range oldMaps {
						if eqMap(newMap, oldMap) {
							found = true
							break
						}
					}
					if !found {
						scopedMap[scopedAttr] = newValue
						return scopedMap, true
					}
				}
			}
		} else {
			for _, newItem := range newValue {
				found := false
				for _, oldItem := range oldSlice {
					if newItem == oldItem {
						found = true
						break
					}
				}
				if !found {
					scopedMap[scopedAttr] = newValue
					return scopedMap, true
				}
			}
		}
		return scopedMap, false
	case interface{}:
		if oldValue, ok := scopedMap[scopedAttr]; !ok || oldValue != newValue {
			scopedMap[scopedAttr] = value
			return scopedMap, true
		}
	}

	return scopedMap, false
}

func (r *replacer) ByValueForItem(scopedSlice []interface{}, value interface{}) ([]interface{}, bool) {
	changed := false
	found := false
	for _, oldValue := range scopedSlice {
		if oldValue == value {
			found = true
			break
		}
	}
	if !found {
		changed = true
		scopedSlice = append(scopedSlice, value)
	}
	return scopedSlice, changed
}

func (r *replacer) ByValueExpressionForItem(scopedMaps []map[string]interface{}, expr filter.Expression, value interface{}) ([]map[string]interface{}, bool) {
	switch newValue := value.(type) {
	case map[string]interface{}:
		changed := false
		newValues := []map[string]interface{}{}
		for _, oldValue := range scopedMaps {
			if !isMatchExpression(oldValue, expr) {
				newValues = append(newValues, oldValue)
			} else {
				if !eqMap(oldValue, newValue) {
					changed = true
					newValues = append(newValues, newValue)
				} else {
					newValues = append(newValues, oldValue)
				}
			}
		}
		return newValues, changed
	default:
		// unexpected input
		return scopedMaps, false
	}
}

func (r *replacer) ByValueExpressionForAttribute(scopedMaps []map[string]interface{}, expr filter.Expression, subAttr string, value interface{}) ([]map[string]interface{}, bool) {
	changed := false
	newValues := []map[string]interface{}{}
	for _, oldValue := range scopedMaps {
		if !isMatchExpression(oldValue, expr) {
			newValues = append(newValues, oldValue)
		} else {
			oldAttrValue, ok := oldValue[subAttr]
			if !ok || oldAttrValue != value {
				changed = true
				oldValue[subAttr] = value
				newValues = append(newValues, oldValue)
			} else {
				newValues = append(newValues, oldValue)
			}
		}
	}
	return newValues, changed
}
