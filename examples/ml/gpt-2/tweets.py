#!/usr/local/bin/python3
from twitter_scraper import get_tweets
import os

for name in os.listdir("/pfs/users/users"):
    with open(os.path.join("/pfs/users/users", name)) as f:
        for user in f:
            user = user.strip()
            with open(os.path.join("/pfs/out", user), "w+") as out:
                for tweet in get_tweets(user):
                    out.write("<|startoftext|> ")
                    out.write(tweet['text'])
                    out.write(" <|endoftext|> ")
