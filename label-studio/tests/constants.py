import os

LABEL_STUDIO_URL = 'http://localhost:8080'
# TODO: Get minikube ip
PACHD_ADDRESS = os.environ.get('PACHD_ADDRESS')
# TODO: remove default
AUTH_TOKEN = os.environ.get('LABEL_STUDIO_USER_TOKEN')