## What has been done.

## Marketplace List
The marketplace list endpoint lists all approved plugins

A [GET] request to /marketplace/plugins will return information of all approved plugins. To request for paginated response, the limit and page should be sent via URL query e.g limit=10&page=1
The response is of this format. The `page`, `limit` and `total` are absent if the request does not include pagination data.
```jsonc
{
  "status": 200,
  "message": "success",
  "data": {
    "limit": 10,      // limit [defaults to 10 if not supplied]
    "page": 1,       // request page
    "plugins": [{}], // list of plugins.
    "total": 1
  }
}
```
To get the next page, increment the page in the response by one.


## Marketplace Search
The marketplace list endpoint lists all approved plugins

A [GET] request to /marketplace/plugins/search?q=query will return information of all approved plugins that match the query term `q` in the `name`, `description`, `category` and `tags` fields of the plugins.
To limit and page should be sent via URL query e.g limit=10&page=1
The response is of this format. The `page`, `limit` and `total` are absent if the request does not include pagination data.
```jsonc
{
  "status": 200,
  "message": "success",
  "data": {
    "limit": 10,      // limit [defaults to 10 if not supplied]
    "page": 1,       // request page
    "plugins": [{}], // list of plugins.
    "total": 1
  }
}
```


## Marketplace GetOne
This [GET] /marketplace/plugins/{id} retreives an approved plugin with the id, and returns data containing the plugin details including the url to install it.

## Marketplace Get Plugin by Template url
This [GET] marketplace/plugins/urls/url?url=<template_url> retreives an approved plugin with the id.
