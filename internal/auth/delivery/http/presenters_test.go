package http

import "testing"

func TestLoginRequestValidation(t *testing.T) {
	cases := []struct {
		name string
		req  loginReq
		want bool
	}{
		{"valid nine digit phone", loginReq{Phone: "090123456", Password: "secret"}, true},
		{"valid eleven digit phone", loginReq{Phone: "09012345678", Password: "secret"}, true},
		{"short phone", loginReq{Phone: "09012345", Password: "secret"}, false},
		{"long phone", loginReq{Phone: "090123456789", Password: "secret"}, false},
		{"phone letters", loginReq{Phone: "09012abcde", Password: "secret"}, false},
		{"short password", loginReq{Phone: "090123456", Password: "short"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.req.validate() == nil; got != tc.want {
				t.Fatalf("validate() success = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoginRequestToInput(t *testing.T) {
	req := loginReq{Phone: "090123456", Password: "secret"}
	input := req.toInput()
	if input.Phone != req.Phone || input.Password != req.Password {
		t.Fatalf("unexpected login input: %+v", input)
	}
}
