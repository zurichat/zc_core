package organizations

import (
	"context"
	"errors"
	"net/http"

	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

type emoji struct {
	UserID string `bson:"user_id"`
	OrgID  string `bson:"organization_id"`
	Emoji  []byte `bson:"emoji"`
}

func EmojiCreator(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, writer)
	}

	filter := make(map[string]interface{})
	filter["_id"] = request.FormValue("user_id")

	res, err := utils.GetMongoDbDoc("users", filter)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, writer)
		return
	}
	if res == nil {
		utils.GetError(errors.New("There is no user here"), http.StatusBadRequest, writer)
		return
	}

	workspaceId := request.FormValue("organization_id")
	workspacesInter := res["workspace_profiles"]
	workspaces := workspacesInter.([]*user.UserWorkspaceProfile)
	for _, wp := range workspaces {
		if wp.OrganizationID == workspaceId {
			break
		}
		utils.GetError(errors.New("There is no you are not here"), http.StatusBadRequest, writer)
		return
	}

	newEmoji := emoji{
		UserID: request.FormValue("user_id"),
		OrgID:  workspaceId,
		Emoji:  getFormFile(writer, request),
	}

	CreateEmoji(newEmoji)

	utils.GetSuccess("Emoji created", nil, writer)
}

func getFormFile(writer http.ResponseWriter, request *http.Request) []byte {
	file, _, err := request.FormFile("emoji")
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, writer)
	}

	buff := make([]byte, 1024)
	_, err = file.Read(buff)
	if err != nil {
		return nil
	}
	return buff
}

func CreateEmoji(e emoji) error {
	collectionName := "emojis"
	ctx := context.Background()
	collection := utils.GetCollection(collectionName)
	_, err := collection.InsertOne(ctx, e)
	if err != nil {
		return err
	}
	return nil
}
