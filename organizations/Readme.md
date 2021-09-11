### What has been done

Adding a plugin to an organization has been implemented.

To add a plugin to an organization, go to the following endpoint with the following data
 [POST] https://api.zuri.chat/organizations/{id}/plugin

where id is the organization id

```json
{
    "plugin_id": "id of the plugin",
    "user_id": "id of the user adding the plugin"
}

```
Every field here is required, else validation error will occur.


Getting all plugins from an organization has been implemented.

To get all plugins from an organization, go to the following endpoint with the following data
 [GET] https://api.zuri.chat/organizations/{id}/plugins

where id is the organization id

Getting a particular plugin from an organization has been implemented.

To add a plugin to an organization, go to the following endpoint with the following data
 [GET] https://api.zuri.chat//organizations/{id}/plugins/{plugin_id}

where id is the organization id