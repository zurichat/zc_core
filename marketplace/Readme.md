## What has been done.


## Marketplace List
The marketplace list endpoint lists all approved plugins

A [GET] request to /marketplace/plugins will return information of all approved plugins


## Marketplace GetOne
This [GET] /marketplace/plugins/{id} retreives an approved plugin with the id, and returns data containing the plugin details including the url to install it.

## Installation of plugins from marketplace
This endpoint at [POST] /marketplace/install takes a json request in the format
```json
{
"organization_id": "xxxx",
"plugin_id": "xxxx",
"user_id": "xxx"
}

```
Successfull installation returns the plugin details, including the template_url which can be displayed by the frontend


To get all plugins in an organization, the [GET] /organizations/{org_id}/plugins endpoint handles that request and returns a list of all plugins for that org

