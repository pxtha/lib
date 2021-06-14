package common

import (
	"github.com/google/uuid"
	"github.com/sendgrid/rest"
	"reflect"
	"testing"
)

func TestTransformString(t *testing.T) {
	type args struct {
		in        string
		uppercase bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "Đ", uppercase: false}, want: "d"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "Â", uppercase: false}, want: "a"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "Á", uppercase: false}, want: "a"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "Ă", uppercase: false}, want: "a"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "À", uppercase: false}, want: "a"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "ơ", uppercase: false}, want: "o"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "ô", uppercase: false}, want: "o"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "ê", uppercase: false}, want: "e"},
		{name: "test", args: struct {
			in        string
			uppercase bool
		}{in: "ư", uppercase: false}, want: "u"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TransformString(tt.args.in, tt.args.uppercase); got != tt.want {
				t.Errorf("TransformString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendRestAPI(t *testing.T) {
	hearder := map[string]string{
		"x-user-id": "7b0a4a54-65f2-40cf-9b16-b2ab044320be",
	}
	type PromotionBody struct {
		Name             *string     `json:"name" valid:"Required"`
		PromotionCode    *string     `json:"promotion_code" valid:"Required"`
		Type             *string     `json:"type" valid:"Required"`
		Value            *float64    `json:"value" valid:"Required"`
		BusinessId       *uuid.UUID  `json:"business_id" valid:"Required"`
	}
	
	res := PromotionBody{
		Name:          StringPointer("a"),
		PromotionCode: nil,
		Type:          nil,
		Value:         nil,
		BusinessId:    nil,
	}
	
	type args struct {
		url        string
		method     rest.Method
		header     map[string]string
		queryParam map[string]string
		bodyInput  interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantBody    string
		wantHeaders map[string][]string
		wantErr     bool
	}{
		// TODO: Add test cases.
		{name: "tét32", args: struct {
			url        string
			method     rest.Method
			header     map[string]string
			queryParam map[string]string
			bodyInput  interface{}
		}{ url: "http://localhost:8081/api/promotion", method: "POST", header:hearder , queryParam: nil, bodyInput:res } , wantBody: "" , wantHeaders:nil , wantErr:true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBody, gotHeaders, err := SendRestAPI(tt.args.url, tt.args.method, tt.args.header, tt.args.queryParam, tt.args.bodyInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendRestAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBody != tt.wantBody {
				t.Errorf("SendRestAPI() gotBody = %v, want %v", gotBody, tt.wantBody)
			}
			if !reflect.DeepEqual(gotHeaders, tt.wantHeaders) {
				t.Errorf("SendRestAPI() gotHeaders = %v, want %v", gotHeaders, tt.wantHeaders)
			}
		})
	}
}