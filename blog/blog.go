package blog

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// An endpoint to list all available blog posts
func GetAllBlogPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogs, err := utils.GetMongoDbDocs(BlogCollectionName, bson.M{"deleted":false})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("success", blogs, response)
}


// An end point to create new blog posts
func CreateBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var blogPost BlogPost
	err := utils.ParseJsonFromRequest(request, &blogPost)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	blogTitle := strings.ToUpper(blogPost.Title)
	
	// confirm if blog title has already been taken
	result, _ := utils.GetMongoDbDoc(BlogCollectionName, bson.M{"title": blogTitle})
	if result != nil {
		utils.GetError(
			errors.New(fmt.Sprintf("blog post with title %s exists!", blogTitle)),
			http.StatusBadRequest, 
			response,
		)
		return
	}

	blogPost.Title = blogTitle
	blogPost.Deleted = false
	blogPost.Likes = 0
	blogPost.Comments = 0
	blogPost.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

	detail, _ := utils.StructToMap(blogPost)

	res, err := utils.CreateMongoDbDoc(BlogCollectionName, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("blog post created", res, response)
}


func ReadBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogID := mux.Vars(request)["blog_id"]
	objID, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	result, err := utils.GetMongoDbDoc(BlogCollectionName, bson.M{"_id": objID, "deleted":false})

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


func UpdateBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogID := mux.Vars(request)["blog_id"]
	objID, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	blogExists, err := utils.GetMongoDbDoc(BlogCollectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusNotFound, response)
		return
	}
	if blogExists == nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusBadRequest, response)
		return
	}

	var blog BlogPost
	if err := utils.ParseJsonFromRequest(request, &blog); err != nil {
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
	} else {
		updateRes, err := utils.UpdateOneMongoDbDoc(BlogCollectionName, blogID, updateFields)
		if err != nil {
			utils.GetError(errors.New("blog post update failed"), http.StatusInternalServerError, response)
			return
		}
		utils.GetSuccess("blog post successfully updated", updateRes, response)
	}
}


func DeleteBlog(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogID := mux.Vars(request)["blog_id"]
	objID, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		utils.GetError(errors.New("invalid blog post ID"), http.StatusBadRequest, response)
		return
	}

	blogExists, err := utils.GetMongoDbDoc(BlogCollectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusNotFound, response)
		return
	}
	if blogExists == nil {
		utils.GetError(errors.New("blog post does not exist"), http.StatusBadRequest, response)
		return
	}

	update := bson.M{"deleted": true, "deleted_at": time.Now()}

	updateRes, err := utils.UpdateOneMongoDbDoc(BlogCollectionName, blogID, update)
	if err != nil {
		utils.GetError(errors.New("blog post could not be deleted"), http.StatusBadRequest, response)
		return
	}

	utils.GetSuccess("blog post successfully deleted", updateRes, response)
}