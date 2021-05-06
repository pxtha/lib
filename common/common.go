package common

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/google/uuid"
)

type Time struct {
	time.Time
}

func (t Time) Value() (driver.Value, error) {
	if !t.IsSet() {
		return "null", nil
	}
	return t.Time, nil
}

func (t *Time) IsSet() bool {
	return t.UnixNano() != (time.Time{}).UnixNano()
}

func Sync(from interface{}, to interface{}) interface{} {
	_from := reflect.ValueOf(from)
	_fromType := _from.Type()
	_to := reflect.ValueOf(to)

	for i := 0; i < _from.NumField(); i++ {
		fromName := _fromType.Field(i).Name
		field := _to.Elem().FieldByName(fromName)
		if !_from.Field(i).IsNil() && field.IsValid() && field.CanSet() {
			fromValue := _from.Field(i).Elem()
			fromType := reflect.TypeOf(fromValue.Interface())
			if fromType.String() == "uuid.UUID" {
				if fromValue.Interface() != uuid.Nil {
					field.Set(fromValue)
				}
			} else if fromType.String() == "string" {
				if field.Kind() == reflect.Ptr {
					tmp := fromValue.String()
					field.Set(reflect.ValueOf(&tmp))
				} else {
					field.Set(fromValue)
				}
			} else if fromType.String() == "service.Time" {
				tmp := fromValue.Interface().(Time)
				if tmp.IsSet() {
					if field.Kind() == reflect.Ptr {
						field.Set(reflect.ValueOf(&tmp))
					} else {
						field.Set(fromValue)
					}
				}
			} else {
				field.Set(fromValue)
			}
		}
	}
	return to
}

func CheckRequireValid(ob interface{}) error {
	validator := validation.Validation{RequiredFirst: true}
	passed, err := validator.Valid(ob)
	if err != nil {
		return err
	}
	if !passed {
		var err string
		for _, e := range validator.Errors {
			err += fmt.Sprintf("[%s: %s] ", e.Field, e.Message)
		}
		return fmt.Errorf(err)
	}
	return nil
}
