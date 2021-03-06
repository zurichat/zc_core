package blog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/utils"
)

// An endpoint to list all available blog posts.
func GetPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogs, err := utils.GetMongoDBDocs(BlogCollectionName, bson.M{"deleted": false})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("success", blogs, response)
}

func GetBlogComments(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	postID := mux.Vars(request)["post_id"]

	result, err := utils.GetMongoDBDoc(BlogCommentsCollectionName, bson.M{"_id": postID})

	if err != nil {
		utils.GetError(errors.New("blog post comments does not exist"), http.StatusNotFound, response)
		return
	}

	if result == nil {
		utils.GetError(errors.New("blog post comments no longer exist"), http.StatusBadRequest, response)
		return
	}

	utils.GetSuccess("success", result, response)
}

// An end point to create new blog posts.
func CreatePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var blogPost Post

	err := utils.ParseJSONFromRequest(request, &blogPost)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	blogTitle := strings.ToTitle(blogPost.Title)

	// confirm if blog title has already been taken
	result, _ := utils.GetMongoDBDoc(BlogCollectionName, bson.M{"title": blogTitle})

	if result != nil {
		utils.GetError(
			fmt.Errorf(fmt.Sprintf("blog post with title %s exists!", blogTitle)),
			http.StatusBadRequest,
			response,
		)

		return
	}

	blogPost.Title = blogTitle
	blogPost.Deleted = false
	blogPost.Likes = 0
	blogPost.Comments = 0
	blogPost.Length = calculateReadingTime(blogPost.Content)
	blogPost.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

	detail, _ := utils.StructToMap(blogPost)

	res, err := utils.CreateMongoDBDoc(BlogCollectionName, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	insertedPostID := res.InsertedID.(primitive.ObjectID).Hex()

	blogPostLikes := Likes{ID: insertedPostID, UsersList: []string{}}
	blogPostLikesMap, _ := utils.StructToMap(blogPostLikes)
	likeDocResponse, err := utils.CreateMongoDBDoc(BlogLikesCollectionName, blogPostLikesMap)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	blogPostComments := BlogsComment{ID: insertedPostID, Comments: []Comment{}}
	blogPostCommentsMap, _ := utils.StructToMap(blogPostComments)

	commentDocResponse, err := utils.CreateMongoDBDoc(BlogCommentsCollectionName, blogPostCommentsMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	ress := []interface{}{res, likeDocResponse, commentDocResponse}

	utils.GetSuccess("blog post created", ress, response)
}

func GetPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	postID := mux.Vars(request)["post_id"]
	objID, err := primitive.ObjectIDFromHex(postID)

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	result, err := utils.GetMongoDBDoc(BlogCollectionName, bson.M{"_id": objID, "deleted": false})

	if err != nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusNotFound, response)
		return
	}

	if result == nil {
		utils.GetError(errors.New("blog post no longer exist"), http.StatusBadRequest, response)
		return
	}

	utils.GetSuccess("success", result, response)
}

func UpdatePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	postID := mux.Vars(request)["post_id"]
	objID, err := primitive.ObjectIDFromHex(postID)

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	blogExists, er := utils.GetMongoDBDoc(BlogCollectionName, bson.M{"_id": objID})

	if er != nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusNotFound, response)
		return
	}

	if blogExists == nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusBadRequest, response)
		return
	}

	var blog Post

	if jsonerr := utils.ParseJSONFromRequest(request, &blog); jsonerr != nil {
		utils.GetError(errors.New("bad update data"), http.StatusUnprocessableEntity, response)
		return
	}

	blog.EditedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

	blogMap, err := utils.StructToMap(blog)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
	}

	updateFields := make(map[string]interface{})

	for key, value := range blogMap {
		if value != "" {
			updateFields[key] = value
		}
	}

	if len(updateFields) == 0 {
		utils.GetError(errors.New("empty/invalid blog input data"), http.StatusBadRequest, response)
		return
	}

	updateRes, err := utils.UpdateOneMongoDBDoc(BlogCollectionName, postID, updateFields)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("blog post successfully updated", nil, response)
}

func DeletePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	postID := mux.Vars(request)["post_id"]
	objID, err := primitive.ObjectIDFromHex(postID)

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	blogExists, err := utils.GetMongoDBDoc(BlogCollectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusNotFound, response)
		return
	}

	if blogExists == nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusBadRequest, response)
		return
	}

	update := bson.M{"deleted": true, "deleted_at": time.Now()}

	updateRes, err := utils.UpdateOneMongoDBDoc(BlogCollectionName, postID, update)
	if err != nil {
		utils.GetError(errors.New("blog post could not be deleted"), http.StatusBadRequest, response)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("blog post successfully deleted", nil, response)
}

func LikeBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var blogLikesDoc Likes

	var userExists bool

	params := mux.Vars(request)
	postID := params["post_id"]
	userID := params["user_id"]
	blogObjID, err := primitive.ObjectIDFromHex(postID)

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"_id": postID}

	blogPostLikes, err := utils.GetMongoDBDoc(BlogLikesCollectionName, filter)
	if err != nil {
		utils.GetError(errors.New("blog post doesn't exist"), http.StatusBadRequest, response)
		return
	}

	blogPostBsonBytes, err := bson.Marshal(blogPostLikes)

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, response)
		return
	}

	_ = bson.Unmarshal(blogPostBsonBytes, &blogLikesDoc)

	for _, value := range blogLikesDoc.UsersList {
		if value == userID {
			userExists = true
		} else {
			userExists = false
		}
	}

	if !userExists {
		updateData := bson.M{"$push": bson.M{"users_list": userID}}

		userLikeResult, err := utils.GenericUpdateOneMongoDBDoc(BlogLikesCollectionName, postID, updateData)

		if err != nil {
			utils.GetError(errors.New("user could not like blog post"), http.StatusBadRequest, response)
			return
		}

		blogPost, err := utils.GenericUpdateOneMongoDBDoc(BlogCollectionName, blogObjID, bson.M{"$inc": bson.M{"likes": 1}})

		if err != nil {
			utils.GetError(errors.New("blog post like count could not be incremented"), http.StatusBadRequest, response)
			return
		}

		if userLikeResult.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
			return
		}

		if blogPost.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
			return
		}

		utils.GetSuccess("user liked successful", blogPost, response)
	} else {
		updateData := bson.M{"$pull": bson.M{"users_list": userID}}

		userLikeResult, err := utils.GenericUpdateOneMongoDBDoc(BlogLikesCollectionName, postID, updateData)

		if err != nil {
			utils.GetError(errors.New("user could not unlike blog post"), http.StatusBadRequest, response)
			return
		}

		blogPost, err := utils.GenericUpdateOneMongoDBDoc(BlogCollectionName, blogObjID, bson.M{"$inc": bson.M{"likes": -1}})

		if err != nil {
			utils.GetError(errors.New("blog post like count could not be decremented"), http.StatusBadRequest, response)
			return
		}

		if userLikeResult.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
			return
		}

		if blogPost.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
			return
		}

		utils.GetSuccess("user disliked successfully", blogPost, response)
	}
}

func CommentBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	params := mux.Vars(request)
	postID := params["post_id"]
	blogObjID, err := primitive.ObjectIDFromHex(postID)

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	var blogComment Comment

	err = utils.ParseJSONFromRequest(request, &blogComment)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	blogComment.ID = primitive.NewObjectID()
	blogComment.CommentAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)
	blogComment.CommentLikes = 0

	blogCommentDoc, err := utils.GetMongoDBDoc(BlogCommentsCollectionName, bson.M{"_id": postID})

	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	if blogCommentDoc == nil {
		utils.GetError(errors.New("blog post comments document doesn't exist"), http.StatusBadRequest, response)
		return
	}

	data, _ := utils.StructToMap(blogComment)

	updateData := bson.M{"$push": bson.M{"comments": data}}

	res, err := utils.GenericUpdateOneMongoDBDoc(BlogCommentsCollectionName, postID, updateData)

	if err != nil {
		utils.GetError(errors.New("comment unsuccessful"), http.StatusBadRequest, response)
		return
	}

	blogPost, err := utils.GenericUpdateOneMongoDBDoc(BlogCollectionName, blogObjID, bson.M{"$inc": bson.M{"comments": 1}})

	if err != nil {
		utils.GetError(errors.New("blog post comment count could not be incremented"), http.StatusBadRequest, response)
		return
	}

	if blogPost.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("comment successful", res, response)
}

func calculateReadingTime(content string) int {
	words := strings.Split(content, " ")
	wordLength := len(words)
	readingTime := wordLength / 200

	return readingTime
}

// SearchBlog returns all posts and aggregates the ones which contain the posted search query in either title or content field.
func SearchBlog(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	blogs := utils.GetCollection("blogs")
	mod := mongo.IndexModel{
		Keys: bson.M{"$**": "text"},
	}

	_, err := blogs.Indexes().CreateOne(context.Background(), mod)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	docs, err := utils.GetMongoDBDocs("blogs", bson.M{"$text": bson.M{"$search": query}})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successful", docs, w)
}

// function to subscribe to a mailing list.
func MailingList(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var mail MailLists

	if err := utils.ParseJSONFromRequest(request, &mail); err != nil {
		utils.GetError(errors.New("bad update data"), http.StatusUnprocessableEntity, response)
		return
	}

	blogMail := strings.ToLower(mail.Email)

	if !utils.IsValidEmail(blogMail) {
		utils.GetError(errors.New("invalid email supplied"), http.StatusBadRequest, response)
		return
	}

	// confirm if email has not already been subscribed
	result, _ := utils.GetMongoDBDoc(BlogMailingList, bson.M{"email": blogMail})
	if result != nil {
		utils.GetError(errors.New("you already subscribed"), http.StatusBadRequest, response)
		return
	}

	mail.Email = blogMail
	mail.SubscribedAt = time.Now()

	detail, _ := utils.StructToMap(mail)

	res, err := utils.CreateMongoDBDoc(BlogMailingList, detail)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("subscribed", res, response)
}
