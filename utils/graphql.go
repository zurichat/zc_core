package utils

import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UserCollectionName         = "users"
	OrganizationCollectionName = "organizations"
)

type GraphQlHandler struct {
	configs *Configurations
}

var ObjectID = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "BSON",
	Description: "The `bson` scalar type represents a BSON Object.",
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.ObjectID:
			return value.Hex()
		case *primitive.ObjectID:
			v := *value
			return v.Hex()
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			id, _ := primitive.ObjectIDFromHex(value)
			return id
		case *string:
			id, _ := primitive.ObjectIDFromHex(*value)
			return id
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			id, _ := primitive.ObjectIDFromHex(valueAST.Value)
			return id
		}
		return nil
	},
})

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Users",
		Fields: graphql.Fields{
			"_id":        &graphql.Field{Type: ObjectID},
			"first_name": &graphql.Field{Type: graphql.String, Description: "First Name"},
			"last_name":  &graphql.Field{Type: graphql.String, Description: "Last Name"},
			"phone":      &graphql.Field{Type: graphql.String, Description: "Phone number"},
			"email":      &graphql.Field{Type: graphql.String, Description: " Email Address"},
			"time_zone":  &graphql.Field{Type: graphql.String, Description: "Time zone"},
			"updated_at": &graphql.Field{Type: graphql.String, Description: "Updated At"},
			"created_at": &graphql.Field{Type: graphql.String, Description: "Created At"},
		},
	},
)

var organizationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Organizations",
		Fields: graphql.Fields{
			"_id":           &graphql.Field{Type: ObjectID},
			"name":          &graphql.Field{Type: graphql.String, Description: "Name"},
			"creator_email": &graphql.Field{Type: graphql.String, Description: "Creator Email"},
			"creator_id":    &graphql.Field{Type: graphql.String, Description: "Creator ID"},
			"admins":        &graphql.Field{Type: graphql.NewList(graphql.String), Description: "Admins"},
			"logo_url":      &graphql.Field{Type: graphql.String, Description: "Logo url"},
		},
	},
)

func loadUsersSchema() *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewList(userType),
		Description: "Get User List",
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			users, err := GetMongoDBDocs(UserCollectionName, bson.M{})
			if err != nil {
				log.Fatal(err.Error())
			}

			return users, nil
		},
	}
}

func loadOrganizationsSchema() *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewList(organizationType),
		Description: "Get Organization list",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			organizations, err := GetMongoDBDocs(OrganizationCollectionName, bson.M{})
			if err != nil {
				log.Fatal(err.Error())
			}
			return organizations, nil
		},
	}
}

var aggregateSchema = graphql.Fields{
	"users":         loadUsersSchema(),
	"organizations": loadOrganizationsSchema(),
}

func (ql *GraphQlHandler) LoadGraphQlSchema() graphql.SchemaConfig {
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: aggregateSchema}
	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
		// Mutation: aggregateMutations,
	}

	return schemaConfig
}

func NewGraphQlHandler(c *Configurations) *GraphQlHandler {
	return &GraphQlHandler{configs: c}
}
