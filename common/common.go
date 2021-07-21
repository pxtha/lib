package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"github.com/google/uuid"
	"github.com/sendgrid/rest"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode"
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

func SendRestAPI(url string, method rest.Method, header map[string]string, queryParam map[string]string, bodyInput interface{}) (body string, headers map[string][]string, err error) {
	request := rest.Request{
		Method:      method,
		BaseURL:     url,
		Headers:     header,
		QueryParams: queryParam,
	}
	if bodyInput != nil {
		bodyData, err := json.Marshal(bodyInput)
		if err != nil {
			return body, headers, err
		}
		request.Body = bodyData
	}
	response, err := rest.Send(request)
	if err != nil {
		return body, headers, err
	} else {
		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusNoContent {
			// parsing error
			r := map[string]interface{}{}
			_ = json.Unmarshal([]byte(response.Body),&r)
			return "", nil, fmt.Errorf("%v", r)
		} else {
			return response.Body, response.Headers, nil
		}
	}
}

// map map[string]string to struct (ex: map mux.Vars to some struct)
func MapStruct(in map[string]string, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	tmp := map[string]interface{}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp2, err := json.Marshal(tmp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(tmp2, out); err != nil {
		return err
	}
	return nil
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func TransformString(in string, uppercase bool) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, err := transform.String(t, in)
	if err != nil {
		logs.Error("Failed to transform %s ", in)
		return ""
	}

	// trim space
	result = strings.ReplaceAll(result, " ", "")
	result = strings.ReplaceAll(result, "Đ", "D")
	result = strings.ReplaceAll(result, "đ", "d")
	if uppercase {
		return strings.ToUpper(result)
	}
	return strings.ToLower(result)
}

func TimeNow() *time.Time {
	t := time.Now()
	return &t
}

func TimePointer(in time.Time) *time.Time {
	return &in
}

func StringPointer(in string) *string {
	return &in
}

func IntPointer(in int) *int {
	return &in
}

func FloatPointer(in float64) *float64 {
	return &in
}
func UUIDPointer(in uuid.UUID) *uuid.UUID {
	return &in
}

type UserHasBusiness struct {
	UserID     uuid.UUID `json:"user_id" gorm:"not null" valid:"Required"`
	BusinessID uuid.UUID `json:"business_id" gorm:"not null" valid:"Required"`
}

type ConsumersKongAResponse struct {
	Data []ConsumersKongA `json:"data"`
}

type ConsumersKongA struct {
	Consumer interface{} `json:"consumer"`
	Id uuid.UUID `json:"id"`
	Group string `json:"group"`
	Tags string  `json:"tags"`
}

func CheckUpdateAnotherUsersData(urlKongGateway string,urlUserHasBusiness string,userIdStr string,creatorId uuid.UUID ,userId uuid.UUID ,businessID uuid.UUID,isCreate bool)error{

	msgError := "Can't update another user's data"

	checkUserRole,err := CheckUserRole(urlKongGateway,userIdStr,"admin")
	if err != nil {
		return err
	}

	if checkUserRole {
		return nil
	}

	if isCreate {
		userHasBusiness,err := GetUserHasBusiness(urlUserHasBusiness,userId.String() ,businessID.String())
		if err != nil {
			return err
		}

		if len(userHasBusiness) == 0{
			logrus.Errorf("Fail to get user has business due to %v", err)
			return fmt.Errorf(msgError)
		}

	}else {
		if creatorId != userId {
			return fmt.Errorf(msgError)
		}
	}

	return nil
}

func GetUserHasBusiness(url string,userId string ,businessID string) (res []UserHasBusiness, err error) {

	param := map[string]string{}
	param["user_id"] = userId
	param["business_id"] = businessID
	body, _, err := SendRestAPI(url,rest.Get, nil, param, nil)
	if err != nil {
		return res, err
	}
	tmp := new(struct{
		Data []UserHasBusiness `json:"data"`
	})
	if err = json.Unmarshal([]byte(body), &tmp); err != nil {
		return res, err
	}
	return tmp.Data, nil
}

func GetRoleUser(url string,userIdStr string) (res []ConsumersKongA, err error) {
	header := make(map[string]string)
	header["x-user-id"] = userIdStr
	body, _, err := SendRestAPI(url, rest.Get, header, nil, nil)
	if err != nil {
		return nil, err
	}
	tmp := new(struct{
		Data []ConsumersKongA `json:"data"`
	})
	if err = json.Unmarshal([]byte(body), &tmp); err != nil {
		return res, err
	}
	return tmp.Data, nil
}

func CheckUserRole(url string,userIdStr string,role string)(bool,error){
	lstConsumer,err := GetRoleUser(url,userIdStr)
	if err != nil {
		return false,err
	}

	for _,consumer := range lstConsumer {
		if strings.Contains(consumer.Group,role) {
			return true,nil
		}
	}

	return false,nil
}