#!/usr/local/bin/python3
import gpt_2_simple as gpt2
import os

os.chdir("/pfs/out")

tweets = [f for f in os.listdir("/pfs/tweets")]

model_name = "117M"
gpt2.download_gpt2(model_name=model_name)


sess = gpt2.start_tf_sess()
gpt2.finetune(sess,
              os.path.join("/pfs/tweets", tweets[0]),
              model_name=model_name,
              steps=1)   # steps is max number of training steps

gpt2.generate(sess)
