package semantic

import "testing"

func BenchmarkIsSQLi_Attack(b *testing.B) {
	input := "1 UNION SELECT * FROM users"
	for b.Loop() {
		IsSQLi(input)
	}
}

func BenchmarkIsSQLi_Normal(b *testing.B) {
	input := "credit union select plan"
	for b.Loop() {
		IsSQLi(input)
	}
}

func BenchmarkIsSQLi_AcceptHeader(b *testing.B) {
	input := "image/avif,image/webp,image/svg+xml,image/*,*/*;q=0.8"
	for b.Loop() {
		IsSQLi(input)
	}
}

func BenchmarkIsXSS_Attack(b *testing.B) {
	input := "<script>alert(1)</script>"
	for b.Loop() {
		IsXSS(input)
	}
}

func BenchmarkIsXSS_Normal(b *testing.B) {
	input := `<div class="container"><p>Hello World</p></div>`
	for b.Loop() {
		IsXSS(input)
	}
}

func BenchmarkIsSQLi_LongInput(b *testing.B) {
	// Simulate a large JSON body
	input := `{"username": "admin", "password": "test123", "email": "user@example.com", "bio": "Software developer & tech enthusiast with 10+ years of experience", "preferences": {"theme": "dark", "language": "en"}}`
	for b.Loop() {
		IsSQLi(input)
	}
}
