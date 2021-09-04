
## Plugins

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
"install_url":  "url to install plugin",
"icon_url": "icon for the plugin"
}

```
Every field here is required.
After a success message is received, The plugin will be approved to appear on the marketplace.
A mock function was created to simulate the approval process, which just waits 10 seconds the approves the plugin.
