package blog

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BlogCollectionName = "blogs"
)


type BlogPost struct {
	ID			primitive.ObjectID     	`bson:"_id,omitempty" json:"id,omitempty"`
	ImageURL	string					`json:"image_url" bson:"image_url"`
	Title		string					`json:"title" bson:"title"`
	Author 		string					`json:"author" bson:"author"`
	Content 	string					`json:"content" bson:"content"`
	Likes 		int						`json:"likes" bson:"likes"`
	Comments 	int						`json:"comments" bson:"comments"`
	Tags		[]string				`json:"tags" bson:"tags"`
	Socials		[]string				`json:"socials" bson:"socials"`
	Length		int						`json:"length" bson:"length"`
	Deleted		bool					`json:"deleted" bson:"deleted"`
	CreatedAt	time.Time				`json:"created_at" bson:"created_at"`
	EditedAt	time.Time				`json:"edited_at" bson:"edited_at"`
	DeletedAt	time.Time				`json:"deleted_at" bson:"deleted_at"`
}


type BlogComment struct {
	ID				string				`bson:"_id,omitempty" json:"id,omitempty"`
	CommentAuthor	string				`json:"comment_author" bson:"comment_author"`
	CommentContent	string				`json:"comment_content" bson:"comment_content"`
	CommentAt 		string				`json:"comment_at" bson:"comment_at"`
	CommentLikes	int					`json:"comment_likes" bson:"comment_likes"`
}

