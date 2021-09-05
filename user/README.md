## User Search Endpoint

Users can be searched for using a "GET" request to the endpoint /users/search/{query}.
The query could be firstname, lastname, display name and email.

```json
{
 "first_name": "xxx",
 "last_name": "xxx",
  "display_name": "xxx",
   "email": "xxx",
}
```
the user json object is returned on successful query.