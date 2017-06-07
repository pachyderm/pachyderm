#!/bin/bash


#cookie_policy=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" | base64 -w 0 |tr '+=/' '-_~' ) 
#signature=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | openssl sha1 -sign cf-keypair-private.pem |base64 -w 0|tr '+=/' '-_~' )
#$keypairid=apkaiqzfqidvcmo6yhxa
keypairid=APKAIQZFQIDVCMO6YHXA

cookie_policy=$(cat etc/deploy/cloudfront/cookie-policy.json | tr -d " \t\n\r" | base64 -w 0 |tr '+=/' '-_~') 
signature=$(cat etc/deploy/cloudfront/cookie-policy.json | tr -d " \t\n\r" | openssl sha1 -sign cf-keypair-private.pem |base64 -w 0 |tr '+=/' '-_~' )

echo "cookie policy [$cookie_policy]"
echo "signature     [$signature]"

curl -v -L \
    -H "Cookie: CloudFront-Policy=$cookie_policy" \
    -H "Cookie: CloudFront-Signature=$signature" \
    -H "Cookie: CloudFront-Key-Pair-Id=$keypairid" \
    http://d2dh50rmvikcjm.cloudfront.net/IMG_3577.PNG

