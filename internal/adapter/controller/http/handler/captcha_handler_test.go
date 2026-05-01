package handler

import "testing"

func TestSafeCaptchaRedirectURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: ""},
		{name: "relative path", raw: "/dashboard?tab=logs", want: "/dashboard?tab=logs"},
		{name: "absolute url", raw: "https://evil.example/phish", want: ""},
		{name: "protocol relative", raw: "//evil.example/phish", want: ""},
		{name: "backslash", raw: `/\evil`, want: ""},
		{name: "relative without slash", raw: "dashboard", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := safeCaptchaRedirectURL(tc.raw); got != tc.want {
				t.Fatalf("safeCaptchaRedirectURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
