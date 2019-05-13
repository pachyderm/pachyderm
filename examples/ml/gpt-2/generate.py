#!/usr/local/bin/python3
import gpt_2_simple as gpt2
import os

models = [f for f in os.listdir("/pfs/train")]

model_dir = os.path.join("/pfs/train", models[0])
os.chdir(model_dir)

sess = gpt2.start_tf_sess()
gpt2.load_gpt2(sess)

out = os.path.join("/pfs/out", models[0])
gpt2.generate_to_file(sess, destination_path=out)
