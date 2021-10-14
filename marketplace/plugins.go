package marketplace

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

// GetAllPlugins returns all approved plugins available in the database.
func GetAllPlugins(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	limit, page := getLimitandPage(query.Get("limit"), query.Get("page"))
    
    opts := options.Find()
    opts.SetLimit(int64(limit))

    if page > 1 {
    	opts.SetSkip(int64(limit * page))
    }

    filter := bson.M{"approved": true}
	ps, err := plugin.FindPlugins(r.Context(), filter , opts)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			utils.GetError(errors.New("no plugin available"), http.StatusNotFound, w)
		default:
			utils.GetError(err, http.StatusNotFound, w)
		}

		return
	}

	utils.GetSuccess("success", utils.M{
		"plugins": ps,
		"page": page,
		"limit": limit,
		"total": utils.CountCollection(r.Context(), "plugins", filter),
	}, w)
}

func getLimitandPage(limit, page string) (int, int) {
	l, _ := strconv.Atoi(limit)
	p, _ := strconv.Atoi(page)

	if p < 1 {
		p = 1
	}

	if l < 1 {
		l = 10
	}
	return l, p	
}

// GetPlugin hanldes the retrieval of a plugin by its id.
func GetPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p, err := plugin.FindPluginByID(r.Context(), id)

	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if !p.Approved {
		utils.GetError(errors.New("plugin is not approved"), http.StatusForbidden, w)
		return
	}

	utils.GetSuccess("success", p, w)
}

// RemovePlugin handles removal of plugins from marketplace.
func RemovePlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	pluginID := mux.Vars(r)["id"]

	pluginExists, err := plugin.FindPluginByID(r.Context(), pluginID)

	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if pluginExists == nil {
		utils.GetError(errors.New("plugin does not exist"), http.StatusBadRequest, w)
		return
	}

	update := bson.M{"approved": false }

	if _, err = utils.UpdateOneMongoDBDoc(plugin.PluginCollectionName, pluginID, update); err != nil {
		utils.GetError(errors.New("plugin removal failed"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("plugin removed", nil, w)
}

// GetPopularPlugins returns all approved plugins available in the database by popularity.
func GetPopularPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.SortPlugins(r.Context(), bson.M{"approved": true }, bson.D{primitive.E{Key: "install_count", Value: -1}})

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			utils.GetError(errors.New("no plugin available"), http.StatusNotFound, w)
		default:
			utils.GetError(err, http.StatusNotFound, w)
		}

		return
	}

	utils.GetSuccess("success", ps, w)
}

// GetPopularPlugins returns all approved plugins available in the database by popularity.
func GetRecomendedPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.SortPlugins(r.Context(), bson.M{"approved": true }, bson.D{primitive.E{Key: "category", Value: 1}})

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			utils.GetError(errors.New("no plugin available"), http.StatusNotFound, w)
		default:
			utils.GetError(err, http.StatusNotFound, w)
		}

		return
	}

	utils.GetSuccess("success", ps, w)
}
