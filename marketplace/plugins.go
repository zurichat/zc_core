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
	opts := options.Find()
	limStr, pgStr := query.Get("limit"), query.Get("page")
	resp := utils.M{}
	filter := bson.M{"approved": true}

	if limStr != "" || pgStr != "" {
		limit, page := getLimitandPage(limStr, pgStr)
		opts.SetLimit(int64(limit)).SetSkip(int64((limit * page) - limit))

		resp["page"], resp["limit"] = page, limit
		resp["total"] = utils.CountCollection(r.Context(), "plugins", filter)
	}

	ps, err := plugin.FindPlugins(r.Context(), filter, opts)

	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	resp["plugins"] = ps

	utils.GetSuccess("success", resp, w)
}

// GetPlugin hanldes the retrieval of a plugin by its id.
func GetPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p, err := plugin.FindPluginByID(r.Context(), id)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if !p.Approved {
		utils.GetError(errors.New("plugin is not approved"), http.StatusForbidden, w)
		return
	}

	utils.GetSuccess("success", p, w)
}

func GetPluginByURL(w http.ResponseWriter, r *http.Request) {
	url:= r.URL.Query().Get("url")
	if url == ""{
		utils.GetError(errors.New("url not supplied"), http.StatusInternalServerError, w)
		return
	}
	
	p, err := plugin.FindPluginByTemplateURL(r.Context(), url)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
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

	update := bson.M{"approved": false}

	if _, err = utils.UpdateOneMongoDBDoc(plugin.PluginCollectionName, pluginID, update); err != nil {
		utils.GetError(errors.New("plugin removal failed"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("plugin removed", nil, w)
}

// GetPopularPlugins returns all approved plugins available in the database by popularity.
func GetPopularPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.SortPlugins(r.Context(), bson.M{"approved": true}, bson.D{primitive.E{Key: "install_count", Value: -1}})

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
	ps, err := plugin.SortPlugins(r.Context(), bson.M{"approved": true}, bson.D{primitive.E{Key: "category", Value: 1}})

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

func Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	filter := bson.M{"$text": bson.M{"$search": q}, "approved": true}
	resp := utils.M{}
	opts := options.Find()

	if query.Get("limit") != "" || query.Get("page") != "" {
		limit, page := getLimitandPage(query.Get("limit"), query.Get("page"))

		opts.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
			SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}).
			SetLimit(int64(limit)).
			SetSkip(int64((limit * page) - limit))

		resp["limit"], resp["page"] = limit, page
		resp["total"] = utils.CountCollection(r.Context(), "plugins", filter)
	}

	docs, err := utils.GetMongoDBDocs("plugins", filter, opts)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	resp["plugins"] = docs

	utils.GetSuccess("success", resp, w)
}

func getLimitandPage(l, p string) (limit, page int) {
	limit, _ = strconv.Atoi(l)
	page, _ = strconv.Atoi(p)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	return
}
