### What has been done

Registration of plugins has been implemented.

To create a plugin, go to the following endpoint with the following data
 [POST] /plugins/register

```json
{
"name": "name of plugin",
"description": "description",
"template_url": "index page of the plugin frontend",
"sidebar_url": "api endpoint to for zuri main to get the plugin sidebar details",
"install_url":  "url for installation",
"developer_name": "whatever",
"developer_email": "whatever@hey.com",
"icon_url": "icon for the plugin"
}

```
Every field here is required, else validation error will occur.
After a success message is received, the plugin is the approved to be listed on marketplace. It takes 10 seconds before the plugin can be listed in the marketplace.
