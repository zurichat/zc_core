package user

import (
	"context"
	"net/http"

	"zuri.chat/zccore/utils"
)

type emoji struct {
	UserFirstName string `bson:"first_name"`
	UserLastName  string `bson:"last_name"`
	Emoji         []byte `bson:"emoji"`
}

func EmojiCreator(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, writer)
	}
	newEmoji := emoji{
		UserFirstName: request.FormValue("first_name"),
		UserLastName:  request.FormValue("last_name"),
		Emoji:         getFormFile(writer, request),
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
