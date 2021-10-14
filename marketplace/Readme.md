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
    "limit": 1,      // limit [defaults to 10 if not supplied]
    "page": 1,       // request page
    "plugins": [{}], // list of plugins.
    "total": 1
  }
}
```
To get the next page, increment the page in the response by one.



## Marketplace GetOne
This [GET] /marketplace/plugins/{id} retreives an approved plugin with the id, and returns data containing the plugin details including the url to install it.

