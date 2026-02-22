package semantic

import "testing"

func TestIsXSS_KnownAttacks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Script tags
		{"basic script tag", "<script>alert(1)</script>", true},
		{"script tag with src", `<script src="evil.js"></script>`, true},
		{"script tag uppercase", "<SCRIPT>alert(1)</SCRIPT>", true},
		{"script tag with space", "<script >alert(1)</script>", true},
		{"closing script tag", "</script>", true},

		// Event handlers
		{"onerror handler", `<img onerror=alert(1)>`, true},
		{"onload handler", `<body onload=alert(1)>`, true},
		{"onclick handler", `<div onclick=alert(1)>`, true},
		{"onmouseover handler", `<a onmouseover=alert(1)>`, true},
		{"onfocus handler", `<input onfocus=alert(1)>`, true},
		{"onblur handler", `<input onblur=alert(1)>`, true},
		{"onsubmit handler", `<form onsubmit=alert(1)>`, true},
		{"onchange handler", `<select onchange=alert(1)>`, true},
		{"event handler with spaces", `<div onclick =alert(1)>`, true},

		// JavaScript protocol
		{"javascript protocol", "javascript:alert(1)", true},
		{"javascript protocol uppercase", "JAVASCRIPT:alert(1)", true},
		{"javascript with space", "javascript :alert(1)", true},
		{"vbscript protocol", `vbscript:msgbox("xss")`, true},

		// Dangerous tags
		{"iframe tag", `<iframe src="evil.com">`, true},
		{"object tag", `<object data="evil.swf">`, true},
		{"embed tag", `<embed src="evil.swf">`, true},
		{"svg with onload", `<svg onload=alert(1)>`, true},
		{"base tag", `<base href="evil.com">`, true},

		// Mixed case / obfuscation
		{"mixed case script", "<ScRiPt>alert(1)</ScRiPt>", true},
		{"javascript colon spaced", "javascript   :alert(1)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsXSS(tt.input)
			if got != tt.want {
				t.Errorf("IsXSS(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsXSS_FalsePositives(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"normal HTML paragraph", "<p>Hello World</p>"},
		{"normal text with script word", "the script was well written"},
		{"normal URL", "https://example.com/path"},
		{"normal class attribute", `<div class="container">`},
		{"normal JSON", `{"event": "click", "handler": "doSomething"}`},
		{"normal link tag reference", "use a script element"},
		{"code snippet mention", "document.getElementById"},
		{"word javascript in text", "I am learning javascript programming"},
		{"email address", "user@example.com"},
		{"empty string", ""},
		{"plain numbers", "12345"},
		{"anchor tag", `<a href="https://example.com">link</a>`},
		{"img tag without events", `<img src="photo.jpg" alt="photo">`},
		{"span tag", `<span class="highlight">text</span>`},
		{"bold tag", `<b>bold text</b>`},
		{"heading tag", `<h1>Title</h1>`},
		{"div with class", `<div class="wrapper"><p>content</p></div>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsXSS(tt.input)
			if got {
				t.Errorf("IsXSS(%q) = true, expected false (false positive)", tt.input)
			}
		})
	}
}
