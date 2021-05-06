package response

type (
	ResponseStatus struct {
		XCode    int    `json:"code" yaml:"code"`
		XStatus  int    `json:"status" yaml:"status"`
		XMessage string `json:"message" yaml:"message"`
	}
)

// New return a new status.
func New(code int, status int, message string) ResponseStatus {
	return ResponseStatus{
		XCode:    code,
		XStatus:  status,
		XMessage: message,
	}
}

func (s ResponseStatus) Error() string {
	return s.XMessage
}

func (s ResponseStatus) Code() int {
	return s.XCode
}

func (s ResponseStatus) Message() string {
	return s.XMessage
}

func (s ResponseStatus) Status() int {
	return s.XStatus
}
