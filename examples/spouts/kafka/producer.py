import os
import csv
import datetime
import time
import uuid

from kafka import KafkaProducer

MAX_ATTEMPTS = 10

def main():
	hostname = os.environ.get("KAFKA_HOSTNAME")
	topic = os.environ.get("KAFKA_TOPIC", "uuids")
	consumer = KafkaProducer(topic, bootstrap_servers=hostname)

	for _ in range(10):
		producer.send(str(uuid.uuid4()))
		producer.flush()

if __name__ == "__main__":
	main()
