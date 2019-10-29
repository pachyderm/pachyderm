#!/usr/bin/python3
import os
import twitterscraper as t

for query in os.listdir("/pfs/queries/"):
    with open(os.path.join("/pfs/queries", query)) as f:
        for q in f:
            q = q.strip()  # clean whitespace
            with open(os.path.join("/pfs/out", query), "w+") as out:
                for tweet in t.query_tweets(q):
                    out.write("<|startoftext|> ")
                    out.write(tweet.text.encode("ascii", "replace").decode("ascii"))
                    out.write(" <|endoftext|> ")
