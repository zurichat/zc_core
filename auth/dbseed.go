package auth

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	// "github.com/nu7hatch/gouuid"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	// "zuri.chat/zccore/user"
	// "zuri.chat/zccore/utils"
)

/*
  Had to add this struct here to implement user login
*/

type User struct {
	ID                  primitive.ObjectID     `bson:"_id"`
	Email               string                 `bson:"email"`
	UserID              string                 `bson:"user_id"`
	FirstName           string                 `bson:"first_name"`
	LastName            string                 `bson:"last_name"`
	Password            string                 `bson:"password"`
	Phone               string                 `bson:"phone"`
	Status              string                 `bson:"status"`
	OrganizationID      string                 `bson:"organization_id"`
	EmailVerificationID string                 `bson:"email_verification_id"`
	PasswordResetIDs    []string               `bson:"password_reset_ids"`
	Settings            map[string]interface{} `bson:"settings"`
	WorkSpaces          map[string]interface{} `bson:"workspaces"`	
	DisplayName         string                 `bson:"display_name" `
	DisplayImage        string                 `bson:"display_image"`
	About               string                 `bson:"about"`
	Timezone            string                 `bson:"timezone"`
	CreatedAt           time.Time              `bson:"created_at"`
	UpdatedAt           time.Time              `bson:"updated_at"`
	DeletedAt           time.Time              `bson:"deleted_at"`
}

func SeedDatabase() {

	db_name, _ :=  os.LookupEnv("DB_NAME")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("CLUSTER_URL")))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	db_collection := client.Database(db_name).Collection("users")
	// usersCollection := database
	// utils.GetMongoDbCollection(db_name, "users")
	 fmt.Println("Collection Type: ", reflect.TypeOf(db_collection))
	// var  user user.User

	// UserID, err := uuid.NewV4();
	// if err != nil {
    //     log.Fatal(err)
    // }

	User_ID := make([]byte, 18)
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    // userInfo := make(map[string]interface{})
	
	bsp, err := bcrypt.GenerateFromPassword([]byte("password"),bcrypt.DefaultCost)
	// user ,
	// )
	if err != nil {
		log.Println(err)
	 }

	//  User := 
	 info := User{
		UserID :  string(User_ID),
		Email :  "go@gmail.com",
		Password :   string(bsp),
		FirstName :  "Zuri",
		LastName :  "Training",
	}

	fmt.Println("Info Type:", reflect.TypeOf(info))

	 result, saveErr := db_collection.InsertOne(ctx, info)
	if saveErr != nil {
		fmt.Println("Insert One Error:", saveErr)
	}
	fmt.Println("InsertOne Result Type:", reflect.TypeOf(result))
	fmt.Println("InsertOne api Result Type:", result)

	// newId := result.InsertedID
	// fmt.Println("InsertOne(), newID", newId)
	// fmt.Println("InsertOne(), newID Type:", reflect.TypeOf(newId))

}

