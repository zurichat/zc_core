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
	PluginCollectionName       = "plugins"
	OrganizationCollectionName = "organizations"
)

type GraphQlHandler struct {
	configs *Configurations
}

// ObjectID for utils.
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

// ********** Users **********.
var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Users",
		Fields: graphql.Fields{
			"_id":        &graphql.Field{Type: ObjectID},
			"first_name": &graphql.Field{Type: graphql.String, Description: "First Name"},
			"last_name":  &graphql.Field{Type: graphql.String, Description: "Last Name"},
			"phone":      &graphql.Field{Type: graphql.String, Description: "Phone number"},
			"email":      &graphql.Field{Type: graphql.String, Description: "Email Address"},
			"time_zone":  &graphql.Field{Type: graphql.String, Description: "Time zone"},
			"updated_at": &graphql.Field{Type: graphql.String, Description: "Updated At"},
			"created_at": &graphql.Field{Type: graphql.String, Description: "Created At"},
		},
	},
)

// ********** Orgnisation **********.
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

// ********** Plugins **********
// PluginType.
var pluginType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Plugins",
		Fields: graphql.Fields{
			"_id":              &graphql.Field{Type: ObjectID},
			"name":             &graphql.Field{Type: graphql.String, Description: "Name"},
			"description":      &graphql.Field{Type: graphql.String, Description: "Description"},
			"developer_name":   &graphql.Field{Type: graphql.String, Description: "DeveloperName"},
			"template_url":     &graphql.Field{Type: graphql.String, Description: "TemplateURL"},
			"sidebar_url":      &graphql.Field{Type: graphql.String, Description: "SidebarURL"},
			"install_url":      &graphql.Field{Type: graphql.String, Description: "InstallURL"},
			"icon_url":         &graphql.Field{Type: graphql.String, Description: "IconURL"},
			"install_count":    &graphql.Field{Type: graphql.Int, Description: "InstallCount"},
			"approved":         &graphql.Field{Type: graphql.Boolean, Description: "Approved"},
			"deleted":          &graphql.Field{Type: graphql.Boolean, Description: "Deleted"},
			"images":           &graphql.Field{Type: graphql.NewList(graphql.String), Description: "Images"},
			"version":          &graphql.Field{Type: graphql.String, Description: "Version"},
			"category":         &graphql.Field{Type: graphql.String, Description: "Category"},
			"tags":             &graphql.Field{Type: graphql.NewList(graphql.String), Description: "Tags"},
			"approved_at":      &graphql.Field{Type: graphql.String, Description: "ApprovedAt"},
			"created_at":       &graphql.Field{Type: graphql.String, Description: "CreatedAt"},
			"updated_at":       &graphql.Field{Type: graphql.String, Description: "UpdatedAt"},
			"deleted_at":       &graphql.Field{Type: graphql.String, Description: "DeletedAt"},
			"sync_request_url": &graphql.Field{Type: graphql.String, Description: "SyncRequestUrl"},
			"queue":            &graphql.Field{Type: MessageModelType, Description: "Queue"},
			"queuepid":         &graphql.Field{Type: graphql.String, Description: "QueuePID"},
		},
	},
)

// MessageModelType ...
var MessageModelType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "MessageModel",
		Fields: graphql.Fields{
			"_id":     &graphql.Field{Type: ObjectID, Description: "Id"},
			"event":   &graphql.Field{Type: graphql.String, Description: "Event"},
			// Please resolve, throw Error: Invalid or incomplete schema, unknown type
			// "message": &graphql.Field{Type: &graphql.Interface{}, Description: "Message"},
		},
	},
)

// Load Organisation Schema.
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

// Load Plugins Schema.
func loadPluginsSchema() *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewList(pluginType),
		Description: "Get Plugins List",
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			plugins, err := GetMongoDBDocs(PluginCollectionName, bson.M{})
			if err != nil {
				log.Fatal(err.Error())
			}

			return plugins, nil
		},
	}
}

// Load User Schema.
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

var aggregateSchema = graphql.Fields{
	"users":         loadUsersSchema(),
	"plugins":       loadPluginsSchema(),
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