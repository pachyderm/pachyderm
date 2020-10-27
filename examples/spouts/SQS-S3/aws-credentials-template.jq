$AWS_ACCESS_KEY_ID | @base64 as $encoded_login |
$AWS_SECRET_ACCESS_KEY | @base64 as $encoded_password |
{
  "apiVersion": "v1",
  "data": {
    "AWS_ACCESS_KEY_ID": $encoded_login,
    "AWS_SECRET_ACCESS_KEY": $encoded_password
  },
  "kind": "Secret",
  "metadata": {
    "name": "aws-credentials"
  },
  "type": "Opaque"
}
