package organizations

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	pluginp "zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

func AddSyncMessage(organization_id string, event string, message interface{}) error {
	plugins, err := GetInstalledPlugins(organization_id)
	if err != nil {
		return err
	}

	err = AddToPluginsQueue(plugins, event, message)
	if err != nil {
		return err
	}

	err = PingPlugins(plugins)
	if err != nil {
		return err
	}

	return nil
}

func PingPlugins(plugins []string) error {
	nw := len(plugins)

	if nw < 1 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(nw)
	wrkchan := make(chan error, nw)

	for _, plgd := range plugins {
		go HandlePingPlugin(plgd, wrkchan, &wg)
	}

	go func() {
		defer close(wrkchan)
		wg.Wait()
	}()

	for n := range wrkchan {
		// println("error ", n)
		if n != nil {
			println("Ping error ", n)
			return n
		}
	}

	return nil
}

func AddToPluginsQueue(plugins []string, event string, message interface{}) error {
	nw := len(plugins)

	if nw < 1 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(nw)
	wrkchan := make(chan error, nw)

	for _, plgd := range plugins {
		go HandleAddingMessage(plgd, event, message, wrkchan, &wg)
	}

	go func() {
		defer close(wrkchan)
		wg.Wait()
	}()

	for n := range wrkchan {
		// println("error ", n)
		if n != nil {
			println("add error ", n)
			return n
		}
	}

	return nil
}

func HandlePingPlugin(plgd string, ch chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	ppID, err := primitive.ObjectIDFromHex(plgd)
	if err != nil {
		ch <- err
		return
	}

	pluginDetails, _ := utils.GetMongoDBDoc(pluginp.PluginCollectionName, bson.M{"_id": ppID})
	if pluginDetails == nil {
		ch <- fmt.Errorf("plugin not found")
		return
	}

	pingUrl := fmt.Sprintf("%v", pluginDetails["sync_request_url"])
	// fmt.Println(pingUrl)
	if pingUrl == "" {
		ch <- fmt.Errorf("no endpoint provided")
		return
	}

	Url, erro := url.Parse(pingUrl)
	if erro != nil {
		ch <- err
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", Url.String(), nil)
	if err != nil {
		ch <- err
		return
	}

	res, err := client.Do(req)
	if err != nil {
		ch <- err
		return
	}
	defer res.Body.Close()
}

func HandleAddingMessage(pluginid, event string, message interface{}, ch chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	ppID, err := primitive.ObjectIDFromHex(pluginid)
	if err != nil {
		ch <- err
		return
	}

	pluginDetails, _ := utils.GetMongoDBDoc(pluginp.PluginCollectionName, bson.M{"_id": ppID})
	if pluginDetails == nil {
		ch <- fmt.Errorf("plugin not found")
		return
	}

	var plugin pluginp.Plugin

	if err = mapstructure.Decode(pluginDetails, &plugin); err != nil {
		ch <- err
		return
	}

	newID := plugin.QueuePID + 1
	newMessage := pluginp.MessageModel{
		Id:      newID,
		Event:   event,
		Message: message,
	}

	updateFields := make(map[string]interface{})
	plugin.Queue = append(plugin.Queue, newMessage)

	updateFields["queue"], updateFields["queuepid"] = plugin.Queue, newID
	_, ee := utils.UpdateOneMongoDBDoc(pluginp.PluginCollectionName, pluginid, updateFields)

	if ee != nil {
		ch <- ee
		return
	}

	ch <- nil
}

func GetInstalledPlugins(organization_id string) ([]string, error) {
	collection := "organizations"

	var org Organization

	objID, err := primitive.ObjectIDFromHex(organization_id)
	if err != nil {
		return nil, err
	}

	orgDetails, _ := utils.GetMongoDBDoc(collection, bson.M{"_id": objID})
	if orgDetails == nil {
		return nil, fmt.Errorf("organization Does not exist")
	}

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(orgDetails)
	err = bson.Unmarshal(bsonBytes, &org)

	if err != nil {
		return nil, err
	}

	pluginSlice := make([]string, 0)

	for _, plgd := range org.OrgPlugins() {
		pluginSlice = append(pluginSlice, fmt.Sprintf("%v", plgd["plugin_id"]))
	}

	return pluginSlice, nil
}
