import os

LABEL_STUDIO_URL = "http://localhost:8080"
## TODO: remove default
AUTH_TOKEN = os.environ.get("LABEL_STUDIO_USER_TOKEN", "abcdef")