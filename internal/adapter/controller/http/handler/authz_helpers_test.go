package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"castellum/internal/usecase/user"

	"github.com/gin-gonic/gin"
)

func TestCanAccessUserResource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user", &user.UserInfoDTO{ID: 42, IsAdmin: false})

	if !canAccessUserResource(c, 42) {
		t.Fatal("expected user to access their own resource")
	}
	if canAccessUserResource(c, 7) {
		t.Fatal("expected non-admin user to be denied for another user's resource")
	}
}

func TestAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		currentUser *user.UserInfoDTO
		wantStatus  int
		wantAbort   bool
	}{
		{
			name:       "missing user",
			wantStatus: http.StatusUnauthorized,
			wantAbort:  true,
		},
		{
			name:        "non admin",
			currentUser: &user.UserInfoDTO{ID: 1, IsAdmin: false},
			wantStatus:  http.StatusForbidden,
			wantAbort:   true,
		},
		{
			name:        "admin",
			currentUser: &user.UserInfoDTO{ID: 1, IsAdmin: true},
			wantStatus:  http.StatusOK,
			wantAbort:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			if tc.currentUser != nil {
				c.Set("user", tc.currentUser)
			}

			AdminMiddleware(c)

			if c.IsAborted() != tc.wantAbort {
				t.Fatalf("unexpected abort state: got %v want %v", c.IsAborted(), tc.wantAbort)
			}
			if recorder.Code != tc.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", recorder.Code, tc.wantStatus)
			}
		})
	}
}
