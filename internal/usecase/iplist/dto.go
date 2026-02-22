package iplist

import "time"

type CreateIpBlackListDto struct {
	Ip        string     `json:"ip" binding:"required"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type IpBlackListDto struct {
	ID        uint       `json:"id"`
	Ip        string     `json:"ip"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateIpWhiteListDto struct {
	Ip        string     `json:"ip" binding:"required"`
	Domain    string     `json:"domain" binding:"required"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateIpWhiteListDto struct {
	Ip        string     `json:"ip"`
	Domain    string     `json:"domain"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type IpWhiteListDto struct {
	ID        uint       `json:"id"`
	Ip        string     `json:"ip"`
	Domain    string     `json:"domain"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type whiteListRule struct {
	IP        string     `json:"ip"`
	Domain    string     `json:"domain"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
