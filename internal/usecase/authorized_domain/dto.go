package authorized_domain

import "time"

type CreateAuthorizedDomainDTO struct {
	Name string `json:"name" binding:"required"`
}

type AuthorizedDomainDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
