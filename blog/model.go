package blog

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BlogCollectionName         = "blogs"
	BlogLikesCollectionName    = "blogslikes"
	BlogCommentsCollectionName = "blogscomments"
	BlogMailingList            = "mailinglist"
)

type BlogPost struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ImageURL  string             `json:"image_url" bson:"image_url"`
	Title     string             `json:"title" bson:"title"`
	Author    string             `json:"author" bson:"author"`
	Content   string             `json:"content" bson:"content"`
	Likes     int                `json:"likes" bson:"likes"`
	Comments  int                `json:"comments" bson:"comments"`
	Tags      []string           `json:"tags" bson:"tags"`
	Socials   []string           `json:"socials" bson:"socials"`
	Length    int                `json:"length" bson:"length"`
	Deleted   bool               `json:"deleted" bson:"deleted"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	EditedAt  time.Time          `json:"edited_at" bson:"edited_at"`
	DeletedAt time.Time          `json:"deleted_at" bson:"deleted_at"`
}

type BlogComment struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	CommentAuthor  string             `json:"comment_author" bson:"comment_author"`
	CommentContent string             `json:"comment_content" bson:"comment_content"`
	CommentAt      time.Time          `json:"comment_at" bson:"comment_at"`
	CommentLikes   int                `json:"comment_likes" bson:"comment_likes"`
}

type BlogsComment struct {
	ID       string        `bson:"_id" json:"_id,omitempty"`
	Comments []BlogComment `bson:"comments" json:"comments,omitempty"`
}

type BlogLikes struct {
	ID        string   `bson:"_id" json:"_id,omitempty"`
	UsersList []string `bson:"users_list" json:"users_list"`
}

type MailLists struct {
	ID              primitive.ObjectID 		`bson:"_id,omitempty" json:".id,omitempty"`
	Email			string					`bson:"email" json:"email"`
	SubscribedAt 	time.Time				`bson:"subscribed_at" json:"subscribed_at"`
}