#!/bin/bash


#cookie_policy=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" | base64 -w 0 |tr '+=/' '-_~' ) 
#signature=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | openssl sha1 -sign cf-keypair-private.pem |base64 -w 0|tr '+=/' '-_~' )
#$keypairid=apkaiqzfqidvcmo6yhxa
keypairid=APKAIQZFQIDVCMO6YHXA

echo "input policy:"
cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" 
echo ""


cookie_policy=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" | base64 -w 0 |tr '+=/' '-_~') 
signature=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" | openssl sha1 -sign cf-keypair-private.pem |base64 -w 0 |tr '+=/' '-_~' )
echo "raw cookie policy [$cookie_policy]"
echo "raw signature     [$signature]"
#signature=$(cat etc/deploy/cloudfront/signed-cookie-policy.json | tr -d " \t\n\r" | base64 -w 0| openssl sha1 -sign cf-keypair-private.pem |base64 -w 0)

# creds from node script
#signature="T2MWGgJKOX-A4Qixi8rG7w4~~vf0LG6uj6Hcv1wulu0aM54XVVoWkJRw9UOWiMtJ6TtIE4v-ymDheDgImhowWBTMDnMBSZ5f0ZmqN0SDXTkF4vMmi~wBercQo75iLYI2T49auwmK1V8bwgtOjTHEdfz66js8DRZbR9gb1wLPj7jqT50XfRnXJpnx16L0yyZK2q5FzlK1b78zXmo59VL7fDioZwYd3dRhtcj8GqRJa1KTHh~aQTKKqS2GABOvBvJVdJlukRmJobYrlEk3J0xMl8gK09ksJUwYS9Xu8vZ~xTptu2xIXBGaoK~9AgWKgj61H4mhfJf-TAyiLzc5tysQnw__"
#cookie_policy="eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cDovL2QzcTJwNHR1YWdreDNwLmNsb3VkZnJvbnQubmV0LyoiLCJDb25kaXRpb24iOnsiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE0OTY3MDk5ODV9fX1dfQ__"


# creds from cookie.rb script
# {"CloudFront-Policy"=>"eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cDovL2QzcTJwNHR1YWdreDNwLmNsb3VkZnJvbnQubmV0L3BhY2gvYmxvY2svNzI0OWE3ZDQxMmRjNGE5N2JiYjg1MWIyNjJhZmI3NTIiLCJDb25kaXRpb24iOnsiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE1MDQyMjQwMDB9fX1dfQ__", "CloudFront-Signature"=>"PEbs4mFLMbcoSjVX-yam9S7HvpZNrugEKrVOVR~l5~n7PU7qK~LtY0iP9n7B7~100MxL~fNEVabEY-eRUubfPj4SiZZbryTeER6vcmQW9J-Aqqxm7z20nArPFrARBHNWnDJb6Tq6cS7aRWPt~ci2PHalGR7iNcXKtauc~XdrW3OENBuAliRbTqcMSVVcR78hsXeYRsuNuZ0xm9fLAHRmhjsHssle-Xr6Gq-UiRuOv7eJunGghSzzoGQzshQ9TMuLe17WXqxvxEjicDSZcLkgc9u~iMN1~pC6rcyruK9Zk-2zHpFtXVLtwTNZPTpMGcW~0HTg00l4-bPaz6BcGBuXOg__", "CloudFront-Key-Pair-Id"=>nil}
#cookie_policy="eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cDovL2QzcTJwNHR1YWdreDNwLmNsb3VkZnJvbnQubmV0L3BhY2gvYmxvY2svNzI0OWE3ZDQxMmRjNGE5N2JiYjg1MWIyNjJhZmI3NTIiLCJDb25kaXRpb24iOnsiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE1MDQyMjQwMDB9fX1dfQ__"
#signature="PEbs4mFLMbcoSjVX-yam9S7HvpZNrugEKrVOVR~l5~n7PU7qK~LtY0iP9n7B7~100MxL~fNEVabEY-eRUubfPj4SiZZbryTeER6vcmQW9J-Aqqxm7z20nArPFrARBHNWnDJb6Tq6cS7aRWPt~ci2PHalGR7iNcXKtauc~XdrW3OENBuAliRbTqcMSVVcR78hsXeYRsuNuZ0xm9fLAHRmhjsHssle-Xr6Gq-UiRuOv7eJunGghSzzoGQzshQ9TMuLe17WXqxvxEjicDSZcLkgc9u~iMN1~pC6rcyruK9Zk-2zHpFtXVLtwTNZPTpMGcW~0HTg00l4-bPaz6BcGBuXOg__"


echo "cookie policy [$cookie_policy]"
echo "signature     [$signature]"

curl -v \
    -H "Cookie: CloudFront-Policy=$cookie_policy" \
    -H "Cookie: CloudFront-Signature=$signature" \
    -H "Cookie: CloudFront-Key-Pair-Id=$keypairid" \
    http://d3q2p4tuagkx3p.cloudfront.net/pach/block/7249a7d412dc4a97bbb851b262afb752
