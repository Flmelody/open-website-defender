package oauth

import (
	"errors"
	"testing"

	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
)

func TestConsentTokenIsOneTimeAndBindsRequest(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	service := newConsentTestService()
	req := validConsentTestRequest()

	token, err := service.CreateConsentToken(req, 42)
	if err != nil {
		t.Fatalf("create consent token: %v", err)
	}
	if token == "" {
		t.Fatal("expected consent token")
	}

	consumed, err := service.ConsumeConsentToken(token, 42)
	if err != nil {
		t.Fatalf("consume consent token: %v", err)
	}
	if *consumed != *req {
		t.Fatalf("consumed request mismatch: got %+v want %+v", consumed, req)
	}

	if _, err := service.ConsumeConsentToken(token, 42); !errors.Is(err, ErrInvalidConsentToken) {
		t.Fatalf("expected second consume to fail, got %v", err)
	}
}

func TestConsentTokenRejectsWrongUser(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	service := newConsentTestService()
	token, err := service.CreateConsentToken(validConsentTestRequest(), 42)
	if err != nil {
		t.Fatalf("create consent token: %v", err)
	}

	if _, err := service.ConsumeConsentToken(token, 7); !errors.Is(err, ErrInvalidConsentToken) {
		t.Fatalf("expected wrong user to be rejected, got %v", err)
	}
	if _, err := service.ConsumeConsentToken(token, 42); !errors.Is(err, ErrInvalidConsentToken) {
		t.Fatalf("expected token to be consumed after wrong-user attempt, got %v", err)
	}
}

func TestCreateConsentTokenValidatesAuthorizeRequest(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	service := newConsentTestService()
	req := validConsentTestRequest()
	req.RedirectURI = "https://attacker.example/callback"

	if token, err := service.CreateConsentToken(req, 42); !errors.Is(err, ErrInvalidRedirectURI) || token != "" {
		t.Fatalf("expected invalid redirect_uri without token, got token=%q err=%v", token, err)
	}
}

func validConsentTestRequest() *AuthorizeRequestDTO {
	return &AuthorizeRequestDTO{
		ResponseType:        "code",
		ClientID:            "client-1",
		RedirectURI:         "https://client.example/callback",
		Scope:               "openid profile",
		State:               "state-1",
		Nonce:               "nonce-1",
		CodeChallenge:       "challenge-1",
		CodeChallengeMethod: "S256",
	}
}

func newConsentTestService() *OAuthService {
	return &OAuthService{
		clientRepo: &fakeOAuthClientRepository{
			client: &entity.OAuthClient{
				ID:           1,
				ClientID:     "client-1",
				ClientSecret: "secret",
				Name:         "Client One",
				RedirectURIs: `["https://client.example/callback"]`,
				Scopes:       "openid profile email",
				Active:       true,
			},
		},
	}
}

type fakeOAuthClientRepository struct {
	client *entity.OAuthClient
}

func (f *fakeOAuthClientRepository) Create(client *entity.OAuthClient) error {
	f.client = client
	return nil
}

func (f *fakeOAuthClientRepository) Update(client *entity.OAuthClient) error {
	f.client = client
	return nil
}

func (f *fakeOAuthClientRepository) Delete(id uint) error {
	f.client = nil
	return nil
}

func (f *fakeOAuthClientRepository) FindByID(id uint) (*entity.OAuthClient, error) {
	if f.client != nil && f.client.ID == id {
		return f.client, nil
	}
	return nil, nil
}

func (f *fakeOAuthClientRepository) FindByClientID(clientID string) (*entity.OAuthClient, error) {
	if f.client != nil && f.client.ClientID == clientID {
		return f.client, nil
	}
	return nil, nil
}

func (f *fakeOAuthClientRepository) List(limit, offset int) ([]*entity.OAuthClient, int64, error) {
	if f.client == nil {
		return nil, 0, nil
	}
	return []*entity.OAuthClient{f.client}, 1, nil
}
