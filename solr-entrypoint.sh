#!/bin/sh
# Create a collection
curl --request POST \
--url http://host.docker.internal:8983/api/collections \
--header 'Content-Type: application/json' \
--data '{
  "create": {
    "name": "teams",
    "numShards": 1,
    "replicationFactor": 1
  }
}'
# Define a schema
curl --request POST \
  --url http://host.docker.internal:8983/api/collections/teams/schema \
  --header 'Content-Type: application/json' \
  --data '{
  "add-field": [
    {"name": "created_at", "type": "pdate", "multiValued": false, "required": true, "stored": true, "indexed": true},
    {"name": "team_user", "type": "string", "multiValued": false, "required": true, "stored": true, "indexed": true},
    {"name": "team_name", "type": "text_general", "multiValued": false, "required": true, "stored": true, "indexed": true},
    {"name": "team_picture", "type": "string", "multiValued": false, "required": false, "stored": true, "indexed": true},
    {"name": "version", "type": "pint", "multiValued": false, "required": true, "stored": true, "indexed": true},
  ],
  "add-copy-field": {"source": "*", "dest": "_text_"}
}'
# Disable automatically to create a field
curl --request POST \
  --url http://host.docker.internal:8983/api/collections/teams/config \
  --header 'Content-Type: application/json' \
  --data '{
    "set-user-property": {"update.autoCreateFields": "false"}
  }'