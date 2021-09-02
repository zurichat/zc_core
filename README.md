# zc_core


### Data read and write for plugins

#### Data Write
Plugins are allowed to write data to the database by calling the /data/write endpoint with any of the POST, PUT, DELETE http methods.
Based on this methods a CREATE, UPDATE or DELETE action will be performed on the database. The plugin would have to provide the following json data body
```json
{
 "plugin_id": "xxx",
 "organization_id": "xxx",
 "collection_name": "mycollection",
 "bulk_write": false,
 "object_id": "xxxx",
 "filter": {},
 "payload": {}
}
```
The plugin_id, organization_id, collection_name fields are important, so it can enable the core api locate the right collection to write data, as the data collections created by plugins are stored seperately for different organizations. 
An internal record is also kept by the core api that validates that this three values are valid, i.e the plugin that is requesting data from this collection for this organization is the one that created it.
This is to prevent other plugins from accessing collections they didn't create.
The `bulk_write` field is a boolean indicating if multiple records are to be inserted, updated or deleted, if it is set to `true`, then `payload` should be an array, if it is false, payload should be a simple plain object.
The `object_id` and `filter` fields are used for updating and deleting data.
If `bulk_write` is to be performed, the `filter` field should be set and should contain the query to be matched for an update, else if performing a single document operation, the `object_id` field should be set instead with the id of the object.
The `payload` contains the actual data the plugin wants to store. The schema is decided by the plugin app. It could be an array of objects or a single object based on if its a bulk_write operation or not.

Once this data is passed, the api performs the operation and sends a response containing the success status and how many documents were successfully written.


#### Data Read

The data read operation occurs at the [GET]  /data/read/{plugin_id}/{collection_name}/{organization_id} endpoint.
Once the api receives this request, it checks the internal record to validate that the plugin with this {plugin_id} is the one that created the {collection_name} for the org with this {organization_id}. Once this is established to be true, then access is granted and the api returns the data requested as an array.
Extra simple mongodb query parameters can be passed as a url query param e.g ?title=this and the api uses it to query the database.


TODO:
- Implement data-write for DELETE, only POST and UPDATE are implemented.

### Marketplace

#### Marketplace List
The marketplace list endpoint lists all approved plugins

A [GET] request to /marketplace/plugins will return the minimal information required to display plugins


#### Marketplace Get Plugin
This [GET] /marketplace/plugins/{id} retreives an approved plugin with the id, and returns data containing the plugin details including the url to install it.


### Plugins

#### Registration
Registration of plugins has been implemented.

To create a plugin, go to the following endpoint with the following data
 [POST] /plugins/register

```json
{
"name": "name of plugin",
"developer_name": "developer",
"developer_email": "dev@developer.mail",
"description": "description",
"template_url": "index page of the plugin frontend",
"sidebar_url": "api endpoint to for zuri main to get the plugin sidebar details",
"install_url":  "not sure what this does, but was in the spec",
"icon_url": "icon for the plugin"
}

```
Every field here is required, else validation error will occur.
After a success message is received, i created a mock fucnction to simulate time it takes to approve a plugin. It takes 10 seconds before the plugin can be listed in the marketplace.

