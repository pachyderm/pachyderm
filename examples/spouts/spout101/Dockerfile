FROM python:3

WORKDIR /app
ADD ./src/requirements.txt requirements.txt
#RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --no-cache-dir -e "git+https://github.com/pachyderm/python-pachyderm.git@spout-auth#egg=python-pachyderm"
ADD ./src/consumer/main.py consumer/main.py
ADD ./src/processor/main.py processor/main.py