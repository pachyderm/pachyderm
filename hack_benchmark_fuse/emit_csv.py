import glob
import pprint
from typing import OrderedDict
zs = OrderedDict()
for d in sorted(glob.glob("output-*")):
    print(f"\n-- {d} --\n")
    xs = OrderedDict()
    cur = None
    for dd in sorted(glob.glob(f"{d}/*.log")):
        for l in open(dd).readlines():
            if l.startswith("TIME:"):
                cur = l.split(":", 1)[1].strip()
                #print(l.strip(), cur)
            if l.strip().startswith("real"):
                xs[cur] = l.strip().split("\t")[1]
            if l.strip().startswith("READ"):
                xs["READ"] = l.split(":", 1)[1].strip()
            if l.strip().startswith("WRITE"):
                xs["WRITE"] = l.split(":", 1)[1].strip()
            if l.strip().startswith("Starting"):
                try:
                    xs["PARAMS"] = l.split(" with ")[1].strip()
                except:
                    pass
    zs[d] = xs

import csv
header = list(zs.values())[0].keys()

with open('data.csv', 'w', newline='') as csvfile:
    writer = csv.writer(csvfile)
    # write the header
    writer.writerow(header)

    for z in zs.values():
        writer.writerow(z.values())
