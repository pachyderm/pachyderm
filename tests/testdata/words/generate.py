#!/usr/bin/python3

import random, time, argparse

def parse_args():
    parser = argparse.ArgumentParser(description='Generate easily-aligned test data for diagnosing Pachyderm\' GCS bug.')
    parser.add_argument('--size', default=int(1e6), type=int, help='The total size of the generated test data')
    return parser.parse_args()


def generate_words():
    words = [ [] for i in range(15) ]
    f = [ l.strip() for l in open("word_list.txt", "r").readlines() ]
    for l in f:
        words[len(l)].append(l)
    return words


def fill(r, size, words):
    # If we're close to the end, finish now
    if size <= 4:
        i = r.randint(0, len(words[size])-1)
        # print("[{}][{}] = {}".format(",".join(words[size]), i, "..."))
        return words[size][i]

    # Pick a random next word, by treating all of 'words' as a single long list
    sz = 0
    for i in range(len(words)):
        if size - i == 1: continue # don't draw from words of this length, so we don't end a line in a space
        sz += len(words[i])

    # Now we have the size of the list we're drawing from. Pick a word
    idx = r.randint(0, sz-1)
    word = ""
    for i in range(len(words)):
        if size - i == 1: continue # not drawing from here
        if idx < len(words[i]):
            word = words[i][idx]
            break
        idx -= len(words[i])

    # Finished the line, we're done
    if len(word) == size:
        return word

    # Didn't finish the line. Add a word and a space, and fill the rest
    remaining = size-len(word)-1
    return word + " " + fill(r, remaining, words[:remaining+1])

def main():
    size=parse_args().size
    words = generate_words()

    out = open("data.txt", "w")
    r = random.Random()
    r.seed(271828)
    for i in range((size+49)//50):
        out.write("{:05}. {}\n".format(i, fill(r, 94, words)))


if __name__ == "__main__":
    main()
