package pointer

import (
	"github.com/google/uuid"
	"time"
)

func TimeNow() *time.Time {
	t := time.Now()
	return &t
}

func String(in string) *string {
	return &in
}

func Int(in int) *int {
	return &in
}

func Float(in float64) *float64 {
	return &in
}

func UUID(in uuid.UUID) *uuid.UUID {
	return &in
}

func Time(in time.Time) *time.Time {
	return &in
}

func BoolPointer(in bool) *bool {
	return &in
}