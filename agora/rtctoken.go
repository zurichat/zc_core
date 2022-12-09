package agora

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

func (a *AgoraHandler) GetRtcToken(w http.ResponseWriter, r *http.Request) {

	// get param values
	channelName, tokentype, uidStr, role := ParseRtcParams(r)
	appId := a.configs.AppId
	appCertificate := a.configs.AppCerificate
	rtcToken, tokenErr := generateRtcToken(appId, appCertificate, channelName, uidStr, tokentype, role)

	if tokenErr != nil {
		utils.GetError(tokenErr, http.StatusBadRequest, w)
		return
	}
	response := AgoraToken{
		Token: rtcToken,
	}
	utils.GetSuccess("Token generated successfully", response, w)

}

func ParseRtcParams(r *http.Request) (channelName, tokentype, uidStr string, role rtctokenbuilder.Role) {
	channelName = mux.Vars(r)["channelName"]
	tokentype = mux.Vars(r)["tokentype"]
	roleStr := mux.Vars(r)["role"]
	r.ParseForm()
	uidStr = r.FormValue("uid")
	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	return channelName, tokentype, uidStr, role
}

func generateRtcToken(appId, appCertificate, channelName, uidStr, tokentype string, role rtctokenbuilder.Role) (string, error) {
	expireTimestamp := time.Now().Add(2 * time.Hour).Unix()
	expireTime := uint32(expireTimestamp)
	if tokentype == "userAccount" {
		rtcToken, err := rtctokenbuilder.BuildTokenWithUserAccount(appId, appCertificate, channelName, uidStr, role, expireTime)
		if err != nil {
			return "", err
		}
		return rtcToken, nil
		// else if tokentype == "uid" {
		// 	uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		// 	// check if conversion fails
		// 	if parseErr != nil {
		// 		err := fmt.Errorf("failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
		// 		return "", err
		// 	}

		// 	uid := uint32(uid64) // convert uid from uint64 to uint 32
		// 	rtcToken, err := rtctokenbuilder.BuildTokenWithUID(appId, appCertificate, channelName, uid, role, expireTime)
		// 	if err != nil {
		// 		return "", err
		// 	}nil
		// 	return rtcToken,

	} else {
		err := fmt.Errorf("failed to generate RTC token for Unknown Tokentype: %s", tokentype)
		return "", err
	}
}
