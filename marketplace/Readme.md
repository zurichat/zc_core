## What has been done.

## Marketplace List
The marketplace list endpoint lists all approved plugins

A [GET] request to /marketplace/plugins?limit=10&page=1 will return information of all approved plugins. This endpoint supports pagination by default, and is limited to 10 if no limit and page values are passed. in the query parameter.
the response is of this format 
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

A [GET] request to /marketplace/plugins/search?q=query&limit=10&page=1 will return information of all approved plugins that match the query term `q` in the `name`, `category` and `tags` fields of the plugins. This endpoint supports pagination by default, and is limited to 10 if no limit and page values are passed. in the query parameter.
the response is of this format 
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

