#!/usr/bin/python3
import twitterscraper as t
import tarfile

while True:
    tweets = t.query_tweets("#PachydermODSC")
    with tarfile.open("/pfs/out", "w") as out:
        for tweet in tweets:
            out.add("/dev/null", "from:%s" % tweet.username)
