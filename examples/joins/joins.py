#!/usr/local/bin/python3

import glob, os

reading_file = glob.glob(os.path.join("/pfs/readings", "*.txt"))[0]
path_file = glob.glob(os.path.join("/pfs/parameters", "*.txt"))[0]
product = 1
for file in [ reading_file, path_file ]:
  sum = 0
  with open(file, 'r') as inp:
    for line in inp:
      sum += int(line)
  product *= sum

name = os.path.split(reading_file)[1] # should be the same as if I did this on path_file
with open("/pfs/out/"+ name, 'w') as out:
  out.write(str(product) + "\n")
