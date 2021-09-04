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

<<<<<<< HEAD
	CreateEmoji(newEmoji)

	utils.GetSuccess("Emoji created", nil, writer)
=======
	utils.GetSuccess("erfdsccc", newEmoji, writer)
>>>>>>> 2baf6599e56be96826d6adcc845800c7246074ad
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

<<<<<<< HEAD
func CreateEmoji(e emoji) error {
=======
func CreateEmoji(e *emoji) error {
>>>>>>> 2baf6599e56be96826d6adcc845800c7246074ad
	collectionName := "emojis"
	ctx := context.Background()
	collection := utils.GetCollection(collectionName)
	_, err := collection.InsertOne(ctx, e)
	if err != nil {
		return err
	}
	return nil
}
