package request

import "time"

type CreateIpBlackListRequest struct {
	Ip        string     `json:"ip" binding:"required"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateIpBlackListRequest struct {
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ListIpBlackListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type CreateIpWhiteListRequest struct {
	Ip        string     `json:"ip" binding:"required"`
	Domain    string     `json:"domain"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateIpWhiteListRequest struct {
	Ip        string     `json:"ip" binding:"required"`
	Domain    string     `json:"domain"`
	Remark    string     `json:"remark"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ListIpWhiteListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
