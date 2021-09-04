## How it works

Create Organization
----------
Users are allowed to create Organization by calling the /organizations endpoint with the POST http methods.
Based on this methods a CREATE action will be performed on the database. The Organization would have to provide the following json data body
```json
{
 "_id": "xxx",
 "name": "xxx",
 "email": "xxx",
 "creator_id": "xxx",
 "plugins": "{[]}",
 "admins": "{[xxx]}",
 "settings": "{}",
 "image_url": "",
 "url": "xxx",
}
```
organization_id, name email creator_id fields are important, so it can enable the core api create the right organization to and stored it seperately organizations in the organizations collections.
The `plugins` this field is an string array containing multiple ids of plugins owned by the organisation.
The `admins` this field is an string array containing multiple ids of users who have authority in organisation.
