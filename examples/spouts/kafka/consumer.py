import os
import csv
import datetime
import time

from kafka import KafkaConsumer

MAX_ATTEMPTS = 10

def main():
	hostname = os.environ.get("KAFKA_HOSTNAME")
	topic = os.environ.get("KAFKA_TOPIC", "uuids")
	consumer = KafkaConsumer(topic, bootstrap_servers=hostname)

	with open("/pfs/out", "w") as f:
		writer = csv.writer(f)

		for message in consumer:
			writer.writerow([str(datetime.datetime.now()), message.value])

if __name__ == "__main__":
	main()
