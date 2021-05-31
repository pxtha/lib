package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	appError interface {
		Error() string
		Code() int
		Message() string

		Status() int
	}

	// BaseResponse is standard status of the app which include code, message, data and meta,...
	BaseResponse struct {
		ResponseStatus
		Data interface{} `json:"data"`
		Meta *string     `json:"meta,omitempty"`
	}

	baseResponse BaseResponse

	// IDResponse is a status helper that has ID
	IDResponse struct {
		ID string `json:"id"`
	}
)

// JSON write status and JSON data to http status writer
func JSON(w http.ResponseWriter, status int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if status == 0 {
		http.Error(w, fmt.Errorf("Utils.Gen ERROR! Make sure you have the tag yaml for GenStatus struct: e.x:` BadRequest   Status `yaml:\"bad_request\"` `").Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

// Error write error, status to http status writer with JSON format: {"code": status, "message": error}
func Error(w http.ResponseWriter, err error, status int) {
	if appError, ok := err.(appError); ok {
		JSON(w, appError.Status(), appError)
		return
	}
	JSON(w, status, map[string]interface{}{
		"status":  status,
		"message": err.Error(),
	})
}

// MarshalJSON implement encoding/json.Marshaler interface.
// It will automatically set AppError to Success if AppError is nil
func (rs BaseResponse) MarshalJSON() ([]byte, error) {
	var v = baseResponse(rs)
	if v.Status() == 0 {
		v.ResponseStatus = ResponseStatus{
			XCode:    1001,
			XStatus:  200,
			XMessage: "Success",
		}
	}
	return json.Marshal(v)
}

