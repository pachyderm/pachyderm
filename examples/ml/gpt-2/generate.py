#!/usr/local/bin/python3
import gpt_2_simple as gpt2
import os

models = [f for f in os.listdir("/pfs/train")]

model_dir = os.path.join("/pfs/train", models[0])
os.chdir(model_dir)

sess = gpt2.start_tf_sess()
gpt2.load_gpt2(sess)

for temp in [0.7 + (x * 0.5) for x in range(3)]:
    out = os.path.join("/pfs/out", models[0]+str(temp))
    gpt2.generate_to_file(sess, destination_path=out,
                          prefix="<|startoftext|>",  truncate="<|endoftext|>",
                          length=280, nsamples=10)
