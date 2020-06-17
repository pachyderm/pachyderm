#!/usr/bin/python3
import gpt_2_simple as gpt2
import os
import argparse

# create parser
parser = argparse.ArgumentParser()

# add arguments to the parser
parser.add_argument("model")
parser.add_argument("steps")


# parse the arguments
args = parser.parse_args()

tweets = [f for f in os.listdir("/pfs/tweets")]

# chdir so that the training process outputs to the right place
out = os.path.join("/pfs/out", tweets[0])
os.mkdir(out)
# chdir to get gpt2 to output where want it to
os.chdir(out)

gpt2.download_gpt2(model_name=args.model)


sess = gpt2.start_tf_sess()
gpt2.finetune(sess,
              os.path.join("/pfs/tweets", tweets[0]),
              model_name=args.model,
              steps=int(args.steps))   # steps is max number of training steps
