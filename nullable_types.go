package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
)

type NullableString struct {
	sql.NullString
}

func (ns NullableString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullableString) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		ns.String = x
	case map[string]interface{}:
		err = json.Unmarshal(data, &ns.NullString)
	case nil:
		ns.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.String", reflect.TypeOf(v).Name())
	}
	ns.Valid = err == nil
	return err
}
