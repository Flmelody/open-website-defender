package request

type CreateIpBlackListRequest struct {
	Ip string `json:"ip" binding:"required"`
}

type ListIpBlackListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type CreateIpWhiteListRequest struct {
	Ip     string `json:"ip" binding:"required"`
	Domain string `json:"domain"`
}

type ListIpWhiteListRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
