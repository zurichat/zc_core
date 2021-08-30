## How it works

Data Write
----------
Plugins are allowed to write data to the database by calling the /data/write endpoint with any of the POST, PUT, DELETE http methods.
Based on this methods a CREATE, UPDATE or DELETE action will be performed on the database. The plugin would have to provide the following json data body
```json
{
 "plugin_id": "xxx",
 "organization_id": "xxx",
 "collection_name": "mycollection",
 "bulk_write": false,
 "object_id": "xxxx",
 "object_ids": [],
 "payload": {}
}
```
The plugin_id, organization_id, collection_name fields are important, so it can enable the core api locate the right collection to write data, as the data collections created by plugins are stored seperately for different organizations. 
An internal record is also kept by the core api that validates that this three values are valid, i.e the plugin that is requesting data from this collection for this organization is the one that created it.
This is to prevent other plugins from accessing collections they didn't create.
The `bulk_write` field is a boolean indicating if multiple records are to be inserted, updated or deleted, if it is set to `true`, then `payload` should be an array, if it is false, payload should be a simple plain object.
The object_id and object_ids fields are used for updating and deleting data.
If `bulk_write` is to be performed, the `object_ids` field should be set and should contain the ids of elements to be updated, else if performing a single document operation, the `object_id` field should be set instead.
The `payload` contains the actual data the plugin wants to store. The schema is decided by the plugin app. It could be an array of objects or a single object based on if its a bulk_write operation or not.

Once this data is passed, the api performs the operation and sends a response containing the success status and how many documents were successfully written.


Data Read
----------
The data read operation occurs at the [GET]  /data/read/{plugin_id}/{collection_name}/{organization_id} endpoint.
Once the api receives this is request, it checks the internal record to validate that the plugin with this {plugin_id} is the one that created the {collection_name} for the org with this {organization_id}. Once this is established to be true, then access is granted and the api returns the data requested as an array.
Extra simple mongodb query parameters can be passed as a url query param e.g ?title=this and the api uses it to query the database.

The link for testing is at https://zccore.herokuapp.com

TODO:
- Extra checks to ensure the plugin and organization records exists in our db before data can be read or written, as at now any plugin_id or org_id can be used to write data.
- Implement bulk_write functionality, only the single object functionality is implemented.
- Implement data-write for UPDATE and DELETE, only POST is implemented.
