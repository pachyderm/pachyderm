# Lifecycle of a Datum
## Introduction
Pachyderm's idea of a "datum" in the context of parallel processing has subtle implications for how data is written to output files. It's important to understand under what conditions data will be overwritten or merged in output files.

There are four basic rules to how Pachyderm will process your data in the pipelines you create.

- Pachyderm will split your input into individual datums _[as you specified](./reference/pipeline_spec.html#the-input-glob-pattern)_ in your pipeline spec
- each datum will be processed independently, using _[the parallelism you specified](../reference/pipeline_spec.html#parallelism-spec-optional)_ in your pipeline spec, with no guarantee of the order of processing
- Pachyderm will merge the output of each pod into each output file, with no guarantee of ordering
- the output files will look as if everything was done in one processing step

If one of your pipelines is written to take all its input, whatever it may be, process it, and put the output into one file, the final output will be one file.  The many datums in your pipeline's input would become one file at the end, with the processing of every datum reflected in that file.

If you write a pipeline to write to multiple files, each of those files may contained merged data that looks as if it were all processed in one step, even if you would have specified the pipeline to process each datum in parallel.

The Pachyderm Pipeline System works with the Pachyderm File System to make sure that files output by each pipeline are merged successfully at the end of every commit.

## Cross and union inputs

When creating pipelines, you can use "union" and "cross" operations to combine inputs.

[Union input](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#union-input) will combine each of the datums in the input repos as one set of datums. The result is that the number of datums processed is the sum of all the datums in each repo.

<img alt="Union input animation: shows how each datum is comprised of each datum in every input to the union." src="../_images/union.gif" width="200px" height="200px"/>

For example, let's say you have two input repos, A and B.  Each of them contain three files with the same names: my-data-file-1.txt, my-data-file-2.txt, and my-data-file-3.txt, but different content.  If you were to cross them in a pipeline, the "input" object in the pipeline spec might look like this.
```
  "input": {
      "union": [
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
If each repo had those three files at the top, there would be six (6) datums overall, which is the sum of the number of input files.  You'd see the following datums, in a random order, in your pipeline as it ran though them.
```
/pfs/A/my-data-file-1.txt

/pfs/A/my-data-file-2.txt

/pfs/A/my-data-file-3.txt

/pfs/B/my-data-file-1.txt

/pfs/B/my-data-file-2.txt

/pfs/B/my-data-file-3.txt
```
One of the ways you can make your code simpler is to use the `name` field for the `pfs` object and give each of those repos the same name.
```
  "input": {
      "union": [
          {
              "pfs": {
                  "name": "C",
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "name": "C",
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
You'd still see the same six datums, in a random order, in your pipeline as it ran though them, but they'd all be in a directory with the same name: C.
```
/pfs/C/my-data-file-1.txt  # from A

/pfs/C/my-data-file-2.txt  # from A

/pfs/C/my-data-file-3.txt  # from A

/pfs/C/my-data-file-1.txt  # from B

/pfs/C/my-data-file-2.txt  # from B

/pfs/C/my-data-file-3.txt  # from B
```
[Cross input](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#cross-input) is essentially a cross-product of all the datums, selected by the globs on the repos you're crossing.  It provides a combination of all the datums to the pipeline that uses it as input, treating each combination as a datum.  

<img alt="Cross input animation: shows how each datum in the pipeline is actually one of the  combinations of every datum across the inputs to the cross" src="../_images/cross.gif" width="200px" height="200px">

There are many examples that show the power of this operator: [Combining/Merging/Joining Data](http://docs.pachyderm.io/en/latest/cookbook/combining.html#combining-merging-joining-data) cookbook and the [Distributed hyperparameter tuning](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter) example are good ones.  

It's important to remember that paths in your input repo will be preserved and prefixed by the repo name to prevent collisions between identically-named files.  For example, let's take those same two input repos, A and B, each of which with the same files as above.  If you were to cross them in a pipeline, the "input" object in the pipeline spec might look like this
```
  "input": {
      "cross": [
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "A"
              }
          },
          {
              "pfs": {
                  "glob": "/*",
                  "repo": "B"
              }
          }
      ]
  }
```
In the case of cross inputs, you can't give the repos being crossed identical names, because of that need to avoid name collisions. If each repo had those three files at the top, there would be nine (9) datums overall, which is every permutation of the input files.  You'd see the following datums, in a random order, in your pipeline as it ran though the nine permutations. 
```
/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-2.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-3.txt
/pfs/B/my-data-file-1.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-2.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-3.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-3.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-2.txt

/pfs/A/my-data-file-1.txt
/pfs/B/my-data-file-3.txt
```
Note that you always see both input directories involved in the cross!
## Output repositories

Every Pachyderm pipeline has an output repository associated with it, with the same name as the pipeline.  [For example](../getting_started/beginner_tutorial.html), an "edges" pipeline would have an "edges" repository you can use as input to other pipelines, like a "montage" pipeline.

Your code, regardless of the pipeline you put it in, should place data in a filesystem mounted under "/pfs/out" and it will appear in the named repository for the current pipeline.

## Appending vs overwriting data in output repositories

The Pachyderm File System keeps track of which datums are being processed in which containers, and makes sure that each datum leaves its unique data in output files.  Let's say you have a simple pipeline, "[wordcount](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count)", that counts references to words in documents by writing the number of occurrences of a word to an output file named for each word in `/pfs/out`, followed by a newline. We intend to process the data by treating each input file as a datum.  We specify the glob in the "wordcount" pipeline json accordingly, something like `"glob": "/*"`. We load a file containing the first paragraph of Charles Dickens's "A Tale of Two Cities" into our input repo, but mistakenly just put the first four lines in the file `tts.txt`.
```
It was the best of times,
it was the worst of times,
it was the age of wisdom,
it was the age of foolishness,
```
In this case, after the pipeline runs on this single datum, `/pfs/out` would contain the files
```
it -> 4\n
was -> 4\n
the -> 4\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
Where "\n" is the newline appended by our "wordcount" code after it outputs the word count. If we were to fix `tts.txt`, either by appending the missing text or replacing it with the entire first paragraph using `pachctl put-file` with the `--overwrite` flag, the file would then look like this
```
It was the best of times,
it was the worst of times,
it was the age of wisdom,
it was the age of foolishness,
it was the epoch of belief,
it was the epoch of incredulity,
it was the season of Light,
it was the season of Darkness,
it was the spring of hope,
it was the winter of despair,
we had everything before us,
we had nothing before us,
we were all going direct to Heaven,
we were all going direct the other way--
in short, the period was so far like the present period, that some of
its noisiest authorities insisted on its being received, for good or for
evil, in the superlative degree of comparison only.
```
We would see each file in the "wordcount" repo overwritten with one line with an updated number. Using our existing examples, we'd see a few of the files replaced with new content
```
it -> 10\n
was -> 10\n
the -> 14\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
The reason that the entire file gets reprocessed, even if we just append to it, is because the entire file is the datum.  We haven't used the `--split` flag combined with the appropriate glob to split it into lots of datums.

What if we have other texts in the pipeline, other datums?  Like the first paragraph of Herman Melville's Moby Dick, put into md.txt?
```
Call me Ishmael. Some years ago—never mind how long precisely—having
little or no money in my purse, and nothing particular to interest me
on shore, I thought I would sail about a little and see the watery part of the world. 
It is a way I have of driving off the spleen and
regulating the circulation. Whenever I find myself growing grim about
the mouth; whenever it is a damp, drizzly November in my soul; whenever
I find myself involuntarily pausing before coffin warehouses, and
bringing up the rear of every funeral I meet; and especially whenever
my hypos get such an upper hand of me, that it requires a strong moral
principle to prevent me from deliberately stepping into the street, and
methodically knocking people's hats off—then, I account it high time to
get to sea as soon as I can. This is my substitute for pistol and ball.
With a philosophical flourish Cato throws himself upon his sword; I
quietly take to the ship. There is nothing surprising in this. If they
but knew it , almost all men in their degree, some time or other,
cherish very nearly the same feelings towards the ocean with me.
```
What happens to our word files?  Will they get overwritten?  Not as long as each input file--`tts.txt` and `md.txt`--is being treated as a separate datum.  You'll see the data in the "wordcount" repo looking something like this:
```
it -> 10\n5\n
was -> 10\n
the -> 14\n7\n
best -> 1\n
worst -> 1\n
of -> 4\n4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
During each job that is run, each distinct datum in Pachyderm will put data in an output file.  If the file shares a name with the files from other datums,  they're merged after processing. This will happen during the appropriately-named _merge_ stage after your pipeline runs. You should not count on the data appearing in a particular order, or the data in a file in an output repo from a particular datum containing the data from the processing of any other datums.  You won't see it in the file until all datums have been processed and the merge is complete, after that pipeline is finished.

What happens if we delete `md.txt`?  The "wordcount" repo would go back to its condition with just `tts.txt`.
```
it -> 10\n
was -> 10\n
the -> 14\n
best -> 1\n
worst -> 1\n
of -> 4\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
What if didn't delete `md.txt`; we appended to it?  Then we'd see the appropriate counts change only on the lines of the files affected by `md.txt`; the counts for `tts.txt` would not change. Let's say we append the second paragraph to `md.txt`:
```
There now is your insular city of the Manhattoes, belted round by
wharves as Indian isles by coral reefs—commerce surrounds it with her
surf. Right and left, the streets take you waterward. Its extreme
downtown is the battery, where that noble mole is washed by waves, and
cooled by breezes, which a few hours previous were out of sight of
land. Look at the crowds of water-gazers there.
```
The "wordcount" repo might now look like this. (We're not using stemmed parser, and "it" is a different word than "its")
```
it -> 10\n6\n
was -> 10\n
the -> 14\n11\n
best -> 1\n
worst -> 1\n
of -> 4\n8\n
times -> 2\n
age -> 2\n
wisdom -> 1\n
foolishness -> 1\n
```
Pachyderm is smart enough to keep track of what changes to what datums affect what downstream results, and only reprocesses and re-merges as needed.

## Summary
To summarize, 

- If you overwrite an input datum in pachyderm with new data, each output datum that can trace to that input datum in downstream pipelines will get updated 
- If your downstream pipeline processes multiple input datums, putting the result a single file, adding or removing an input datum will only remove its effect from that file.  The effect of the other datums will still be seen in that file.

You can see this in action in the [word count example](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count) in the Pachyderm git repo.



