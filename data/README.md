## How it works

Data Write (Create and Update)
----------
Plugins are allowed to write data to the database by calling the /data/write endpoint with any of the POST, PUT, http methods.
Based on this methods a CREATE, UPDATE action will be performed on the database. The plugin would have to provide the following json data body
```json
{
 "plugin_id": "xxx",
 "organization_id": "xxx",
 "collection_name": "mycollection",
 "bulk_write": false,
 "object_id": "xxxx",
 "filter": {},
 "payload": {},
 "raw_query": {}
}
```
The plugin_id, organization_id, collection_name fields are important, so it can enable the core api locate the right collection to write data, as the data collections created by plugins are stored seperately for different organizations. 
An internal record is also kept by the core api that validates that this three values are valid, i.e the plugin that is requesting data from this collection for this organization is the one that created it.
This is to prevent other plugins from accessing collections they didn't create.
The `bulk_write` field is a boolean indicating if multiple records are to be inserted, updated or deleted, if it is set to `true`, then `payload` should be an array, if it is false, payload should be a simple plain object.
The object_id and filter fields are used for updating and deleting data.
If `bulk_write` is to be performed, the `filter` field should be set and should contain the query to be matched for an update, else if performing a single document operation, the `object_id` field should be set instead with the id of the object.
The `payload` contains the actual data the plugin wants to store. The schema is decided by the plugin app. It could be an array of objects or a single object based on if its a bulk_write operation or not.

The `raw_query` field allows the user to perform raw mongodb queries when updating like using aggregation operations, $addToSet etc etc and should only be used if you know what you are doing.

Once this data is passed, the api performs the operation and sends a response containing the success status and how many documents were successfully written.


Data Read
----------
The data read operation occurs at the [GET]  /data/read/{plugin_id}/{collection_name}/{organization_id} endpoint.
Once the api receives this request, it checks the internal record to validate that the plugin with this {plugin_id} is the one that created the {collection_name} for the org with this {organization_id}. Once this is established to be true, then access is granted and the api returns the data requested as an array.
Extra simple mongodb query parameters can be passed as a url query param e.g ?title=this and the api uses it to query the database.
To find an item by id, pass in the query parameter `id` or `_id` with the appropriate value . A single document/object is returned instead of a list if the item is found.

An alternative implementation is exposed at the [POST] /data/read endpoint. The request payload expected is of the following format. 
```jsonc
{
    "plugin_id": "",
    "collection_name": "",
    "organization_id": "",
    "object_id": "",
    "object_ids": [], // useful if you want to get results for a list of ids.
    "filter": {}, // can be used in place of id to match documents by other fields.
    "options": {} // contains query results modifiers.
}

```
the `filter` object should contain the field(s) you want to match and their values. The `options` is an object with the following valid properties:

- limit (integer): To set a limit on the results to be returned.
- skip (integer): to set an offset value for the results.
- sort (Object): field to sort by in descending (-1) or ascending order (1) e.g `"sort": {"name": 1}`
- projection (Object): field to include (1) or exclude (0) in the results e.g `"projection": {"field": 1}`


Delete Data
-----------
To delete data, a POST request is made to /data/delete

```json
{
 "plugin_id": "xxx",
 "organization_id": "xxx",
 "collection_name": "mycollection",
 "bulk_delete": false,
 "object_id": "xxxx",
 "filter": {},
```

The `bulk_delete` and `filter` properties are used to delete multiple records. `filter` will contain the query to be matched and `bulk_delete` must be set to `true` to use this filter property.


List Data Collections
---------------------
Plugins can now request to see a list of collections they have created. 

The a GET request to `/data/collections/<plugin_id>` will return a record of collections created by the plugin.
while a request to `/data/collections/<plugin_id>/<org_id>` will return a record of collections a plugin has created for a particular organization

