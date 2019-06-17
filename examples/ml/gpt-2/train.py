#!/usr/bin/python3
import gpt_2_simple as gpt2
import os


tweets = [f for f in os.listdir("/pfs/tweets")]

# chdir so that the training process outputs to the right place
out = os.path.join("/pfs/out", tweets[0])
os.mkdir(out)
# chdir to get gpt2 to output where want it to
os.chdir(out)

model_name = "345M"
gpt2.download_gpt2(model_name=model_name)


sess = gpt2.start_tf_sess()
gpt2.finetune(sess,
              os.path.join("/pfs/tweets", tweets[0]),
              model_name=model_name,
              steps=1000)   # steps is max number of training steps
