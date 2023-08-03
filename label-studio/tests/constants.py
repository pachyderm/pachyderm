import os

LABEL_STUDIO_URL = 'http://localhost:8080'
# TODO: Get minikube ip
PACHD_ADDRESS = 'host.docker.internal:30650'
# TODO: remove default
AUTH_TOKEN = os.environ.get('LABEL_STUDIO_USER_TOKEN', 'abcdef')