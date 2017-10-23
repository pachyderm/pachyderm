#!/bin/bash

for play in hamlet julius_caesar macbeth othello; do
  curl http://shakespeare.mit.edu/${play}/full.html >>shakespeare_plays_1mb.txt
done


tmp=$(mktemp --tmpdir=.)
cp shakespeare_plays_1mb.txt ${tmp}
for i in $(seq 100); do
  cat ${tmp} >> shakespeare_plays_100mb.txt;
done
rm ${tmp}

gsutil cp shakespeare_plays_1mb.txt gs://shakespeare_plays
gsutil cp shakespeare_plays_100mb.txt gs://shakespeare_plays
