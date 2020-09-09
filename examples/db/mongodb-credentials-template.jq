$uri | @base64 as $encoded_uri |
$username | @base64 as $encoded_username |
$password | @base64 as $encoded_password |
$db | @base64 as $encoded_db |
$collection | @base64 as $encoded_collection |    
{
  "apiVersion": "v1",
  "data": {
      "uri": $encoded_uri,
      "username": $encoded_username,
      "password": $encoded_password,
      "db": $encoded_db,
      "collection": $encoded_collection
  },
  "kind": "Secret",
  "metadata": {
    "name": "mongosecret"
  },
  "type": "Opaque"
}
