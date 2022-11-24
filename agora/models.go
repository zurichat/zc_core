package agora

import "zuri.chat/zccore/utils"

type AgoraToken struct {
	Token string `json:"token"`
}

type AgoraHandler struct {
	configs *utils.Configurations
}

func NewAgoraHandler(c *utils.Configurations) *AgoraHandler {
	return &AgoraHandler{configs: c}
}
