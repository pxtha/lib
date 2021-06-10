package common

import "testing"

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