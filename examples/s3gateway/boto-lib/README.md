# Boto Python Library

Demonstrates interaction with the s3 gateway through [boto3](https://github.com/boto/boto3).

To run the example:

1) [Setup a cluster](https://docs.pachyderm.com/latest/getting_started/local_installation/).
2) Go through [the OpenCV example](https://docs.pachyderm.com/latest/examples/examples/#opencv-edge-detection).
3) Setup a virtualenv: `virtualenv -v venv -p python3`.
4) Install dependencies: `. venv/bin/activate && pip install -r requirements.txt`.
5) If the cluster isn't running on `localhost`, run `pachctl port-forward`.
5) Run the example: `. venv/bin/activate && ./main.py`.
