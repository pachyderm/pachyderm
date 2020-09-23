$IMAP_LOGIN | @base64 as $encoded_login |
$IMAP_PASSWORD | @base64 as $encoded_password |
{
  "apiVersion": "v1",
  "data": {
    "IMAP_LOGIN": $encoded_login,
    "IMAP_PASSWORD": $encoded_password
  },
  "kind": "Secret",
  "metadata": {
    "name": "imap-credentials"
  },
  "type": "Opaque"
}
