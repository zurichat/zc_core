#### Data package
This package facilitates transfer of data between plugins and core app database

- There is an endpoint to read data
- Another one to write data to the database

Only reading from the database is implemented

###### How READ operation works
A plugin sends a get request to the /data/read endpoint, passing in necessary information about the data it wants to read

e.g:
* GET /read/<plugin_id>/<collection_name>[?query_param] :-> returns collection_data to plugin

The plugin must have previously requested the collection be made so that the collection is available in the database and a PluginData record is kept to manage access to this collection.

Since every collection in the db should have a unique name and be created by a plugin, we can easily use <collection_name> to find a PluginData item.
We then compare the owner_plugin_id field in the PluginData with <plugin_id> to ensure plugins can access only collectionscreated by them.
Then we perform a read operation on the specified collection and return all results to the plugin to process
If a query parameter e.g `_id=xxxx` is passed, it can be used to filter results at db level before returning it to plugin.as JSON
