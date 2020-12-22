#!/usr/local/bin/python3
import python_pachyderm
import os
import hashlib
from random import randint
from time import sleep

# Emulates the reception of messages from a third party messaging system or queue (such as AWS SQS, Kafka, Google Pub/Sub etc...)


def receive_message():
    # Emulates a network response time to poll new messages
    print("receiving")
    sleep(randint(10, 30))
    print("emiting")
    # Creates a random string of 1KB
    random1 = os.urandom(1024)
    random2 = os.urandom(2048)
    return (random1, random2)

# Polls data from a third party messaging system or queue and push them to a Pachyderm repo in a transaction.


def polling_consumer():
    print("polling")
    while True:
        # Polls queue
        msgs = receive_message()
        print("message: " + str(msgs[0]))
        if msgs:
            print("connecting to pachd")
            client = python_pachyderm.Client(host="pachd", port=650)
            print(client.health())
            print("connected")
            with client.commit('spout', 'master') as c:
                print("creating commit")
                for msg in msgs:
                    # hash the file to assign unique name
                    filename = hashlib.sha256(msg).hexdigest() + ".txt"
                    print("adding file" + filename)
                    client.put_file_bytes(c, filename, msg)


if __name__ == '__main__':
    polling_consumer()
