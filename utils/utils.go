package utils

import (
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"reflect"
	"time"
)

// Custom decode hook for dbtype.Date to time.Time conversion
func dateToTimeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from == reflect.TypeOf(dbtype.Date{}) && to == reflect.TypeOf(time.Time{}) {
		return data.(dbtype.Date).Time(), nil
	}
	return data, nil
}

// Reusable decoder function
func DecodeToStruct(input map[string]interface{}, output interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook:       dateToTimeHook,
		Result:           output,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
