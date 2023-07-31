print("starting tests...")
import requests
import os

LABEL_STUDIO_HOST = "http://localhost:8080"
## TODO: remove default
AUTH_TOKEN = os.environ.get("LABEL_STUDIO_USER_TOKEN", "abcdef")

def test_begin():
    x = requests.get(f"{LABEL_STUDIO_HOST}/api/current-user/whoami", headers={'Authorization': f'Token {AUTH_TOKEN}'})
    print(x.json())

    assert x.ok