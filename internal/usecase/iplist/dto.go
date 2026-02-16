package iplist

import "time"

type CreateIpBlackListDto struct {
	Ip string `json:"ip" binding:"required"`
}

type IpBlackListDto struct {
	ID        uint      `json:"id"`
	Ip        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateIpWhiteListDto struct {
	Ip     string `json:"ip" binding:"required"`
	Domain string `json:"domain" binding:"required"`
}

type UpdateIpWhiteListDto struct {
	Ip     string `json:"ip"`
	Domain string `json:"domain"`
}

type IpWhiteListDto struct {
	ID        uint      `json:"id"`
	Ip        string    `json:"ip"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}
