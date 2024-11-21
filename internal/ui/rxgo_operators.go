package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func SafeTransform(fn func(value interface{}, root map[string]interface{}) (interface{}, error), path string) func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		// Ensure input is a map (the combined structure from CombineLatest)
		root, ok := input.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("input is not a valid combined map: %v", input)
		}

		// Create a deep copy of the root structure to ensure immutability
		rootCopy, err := deepCopy(root)
		if err != nil {
			return nil, fmt.Errorf("error creating a deep copy: %w", err)
		}

		// Navigate to the specified path in the copied structure
		parts := strings.Split(path, ".")
		target, err := navigateRecursive(rootCopy, parts)
		if err != nil {
			return nil, fmt.Errorf("error navigating to path %s: %w", path, err)
		}

		// Apply the transformation function with the resolved value and the root structure
		transformed, err := fn(target, rootCopy)
		if err != nil {
			return nil, fmt.Errorf("error applying transformation at path %s: %w", path, err)
		}

		// Update the copied structure with the transformed value
		updatedRoot, err := updateRecursive(rootCopy, parts, transformed)
		if err != nil {
			return nil, fmt.Errorf("error updating data at path %s: %w", path, err)
		}

		return updatedRoot, nil
	}
}

// deepCopy creates a deep copy of a map to ensure immutability.
func deepCopy(data map[string]interface{}) (map[string]interface{}, error) {
	copy := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively copy maps
			copiedValue, err := deepCopy(v)
			if err != nil {
				return nil, err
			}
			copy[key] = copiedValue
		case []interface{}:
			// Copy slices
			copiedSlice := make([]interface{}, len(v))
			for i, item := range v {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					copiedItem, err := deepCopy(nestedMap)
					if err != nil {
						return nil, err
					}
					copiedSlice[i] = copiedItem
				} else {
					copiedSlice[i] = item
				}
			}
			copy[key] = copiedSlice
		default:
			// Copy primitive types
			copy[key] = v
		}
	}
	return copy, nil
}

// Recursively navigates a data structure to retrieve the value at the specified path.
func navigateRecursive(data interface{}, parts []string) (interface{}, error) {
	if len(parts) == 0 {
		return data, nil
	}

	part := parts[0]

	switch v := data.(type) {
	case map[string]interface{}:
		// Handle map keys
		key, arrayIndex, err := parseKeyAndIndex(part)
		if err != nil {
			return nil, fmt.Errorf("invalid key or index in path: %s", part)
		}
		value, ok := v[key]
		if !ok {
			return nil, fmt.Errorf("key %s not found", key)
		}
		if arrayIndex != nil {
			// Handle arrays within maps
			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("key %s is not an array", key)
			}
			if *arrayIndex < 0 || *arrayIndex >= len(arr) {
				return nil, fmt.Errorf("index %d out of bounds for key %s", *arrayIndex, key)
			}
			value = arr[*arrayIndex]
		}
		return navigateRecursive(value, parts[1:])

	case []interface{}:
		// Handle array indices
		arrayIndex, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", part)
		}
		if arrayIndex < 0 || arrayIndex >= len(v) {
			return nil, fmt.Errorf("index %d out of bounds", arrayIndex)
		}
		return navigateRecursive(v[arrayIndex], parts[1:])

	default:
		return nil, fmt.Errorf("cannot navigate %s: not a map or array", part)
	}
}

// Recursively updates a value in the data structure at the specified path.
func updateRecursive(data interface{}, parts []string, value interface{}) (interface{}, error) {
	if len(parts) == 0 {
		return value, nil
	}

	part := parts[0]

	switch v := data.(type) {
	case map[string]interface{}:
		// Handle map keys
		key, arrayIndex, err := parseKeyAndIndex(part)
		if err != nil {
			return nil, fmt.Errorf("invalid key or index in path: %s", part)
		}
		current, ok := v[key]
		if !ok {
			return nil, fmt.Errorf("key %s not found", key)
		}
		if arrayIndex != nil {
			// Handle arrays within maps
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("key %s is not an array", key)
			}
			if *arrayIndex < 0 || *arrayIndex >= len(arr) {
				return nil, fmt.Errorf("index %d out of bounds for key %s", *arrayIndex, key)
			}
			arr[*arrayIndex], err = updateRecursive(arr[*arrayIndex], parts[1:], value)
			if err != nil {
				return nil, err
			}
			v[key] = arr
		} else {
			// Navigate further down the map
			v[key], err = updateRecursive(current, parts[1:], value)
			if err != nil {
				return nil, err
			}
		}
		return v, nil

	case []interface{}:
		// Handle array indices
		arrayIndex, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", part)
		}
		if arrayIndex < 0 || arrayIndex >= len(v) {
			return nil, fmt.Errorf("index %d out of bounds", arrayIndex)
		}
		v[arrayIndex], err = updateRecursive(v[arrayIndex], parts[1:], value)
		if err != nil {
			return nil, err
		}
		return v, nil

	default:
		return nil, fmt.Errorf("cannot update %s: not a map or array", part)
	}
}

// Parses a key and optional array index from a path component.
func parseKeyAndIndex(part string) (key string, index *int, err error) {
	if strings.Contains(part, "[") && strings.Contains(part, "]") {
		openBracket := strings.Index(part, "[")
		closeBracket := strings.Index(part, "]")
		key = part[:openBracket]
		indexStr := part[openBracket+1 : closeBracket]
		idx, err := strconv.Atoi(indexStr)
		if err != nil {
			return "", nil, fmt.Errorf("invalid array index: %s", indexStr)
		}
		return key, &idx, nil
	}
	return part, nil, nil
}
