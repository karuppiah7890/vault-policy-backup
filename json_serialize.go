package main

import "encoding/json"

func toJSON(value interface{}) ([]byte, error) {
	valueAsJSON, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return valueAsJSON, nil
}
