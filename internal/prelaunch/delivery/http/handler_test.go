package http

import "testing"

func TestProjectRequestValidation(t *testing.T) {
	cases := []struct {
		name string
		req  projectRequest
		want bool
	}{
		{"https website", projectRequest{Name: "Nova Chain", WebsiteURL: "https://nova.example"}, true},
		{"http website", projectRequest{Name: "Nova Chain", WebsiteURL: "http://localhost:3000"}, true},
		{"short name", projectRequest{Name: "N", WebsiteURL: "https://nova.example"}, false},
		{"missing host", projectRequest{Name: "Nova Chain", WebsiteURL: "https:///missing-host"}, false},
		{"unsafe scheme", projectRequest{Name: "Nova Chain", WebsiteURL: "javascript:alert(1)"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.req.valid(); got != tc.want {
				t.Fatalf("valid() = %v, want %v", got, tc.want)
			}
		})
	}
}
