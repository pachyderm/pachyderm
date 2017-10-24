#!/usr/bin/python3

import sys, tempfile, os

def ln(s):
    idx = s.find(".")
    if idx < 0:
        return -1
    return int(s[:idx])

def fill_out_gap(itr, bigf_pre, smallf_pre, bigf, smallf, out):
    out.write(bigf_pre)
    big_n, small_n = bigf.readline(), smallf.readline()
    if big_n == "": return
    out.write(big_n)
    if small_n == "":
        out.writelines(bigf.readlines())

    bad = "" # Total hack. This holds a line in small in case its line number was interrupted
    start_ln, end_ln = ln(big_n), ln(small_n)
    if end_ln == -1:
        bad = small_n
        small_n  = smallf.readline()
        end_ln = ln(small_n)
    for i in range(start_ln+1, end_ln-1):
        out.write(bigf.readline())

    big_n = bigf.readline()
    suffix = ""
    #print("trying to find the shared suffix of \"{}\" and \"{}\"".format(big_n, smallf_pre))
    for i in range(min(len(big_n), len(smallf_pre))):
        if big_n[-1-i] != smallf_pre[-1-i]:
            break
        suffix += smallf_pre[-1-i]
        #print("suffix = {}".format(suffix[:1:-1]))
    suffix = suffix[::-1]
    #print("{}[{}]".format(big_n[:-len(suffix)], big_n[-len(suffix):]))
    out.write(big_n[:-len(suffix)])

    if bad != "":
        out.write(bad)

    new_out = open("{:02d}.shared.txt".format(itr), "w")
    new_out.write(suffix)
    big_n = bigf.readline()
    fill_out_same_text(itr+1, big_n, small_n, bigf, smallf, new_out)


def fill_out_same_text(itr, big_n, small_n, bigf, smallf, out):
    while True:
        if small_n == "":
            out.write(big_n)
            out.writelines(bigf.readlines())
            return
        if big_n == small_n:
            out.write(big_n)
        else:
            break
        big_n, small_n = bigf.readline(), smallf.readline()

    # Determine which parts of the line are the same, and then jump to finding
    # the gap once they diverge
    same = ""
    for i in range(min(len(big_n), len(small_n))):
        if big_n[i] == small_n[i]:
            same += big_n[i]
        else:
            break
    out.write(same)
    out.close()

    new_out = open("{:02d}.gap.txt".format(itr), "w")
    fill_out_gap(itr+1, big_n[len(same):], small_n[len(same):], bigf, smallf, new_out)


l, r = open(sys.argv[1], "r"), open(sys.argv[2], "r")
tempdir = tempfile.mkdtemp(prefix="file_chunks.", dir=".")
os.chdir("./{}".format(tempdir))
new_out = open("00.shared.txt", "w")
fill_out_same_text(1, l.readline(), r.readline(), l, r, new_out)
