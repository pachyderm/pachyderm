Say you have a cross pipeline (called cross-pipe) with three input branches A, B and C.

Each of these branches has had commits on it at various times, each new commit triggering a new job from the pipeline.

Now suppose you have commits a1, a2, a3 and a4 on branch A, b1 and b2 on branch B, and c1, c2 and c3 on branch C.

Maybe now you are curious to see the result of the pipeline with the combination of data after a4, b1 and c2 were committed.
But none of the output commits were triggered on this particular combination.

In order to get the result of this, you can run `pachctl run pipeline cross-pipe a4 b1 c2`.
This will create a new job which will create a commit on the pipeline's output branch with the result of that combination of data.

Since a4 is the head of branch A, we could also have done `pachctl run pipeline cross-pipe c2 b1`, which would give the same result.
Pachyderm is smart enough to automatically use the head for any branches that didn't have a commit specified. 
Note also that the order you put the commits in doesn't matter, Pachyderm knows how to match them to the right place.
You also (as always) can reference the head commit of a branch by using the branch name: `pachctl run pipeline cross-pipe A b1 c2`.

This means that if you wanted to run the pipeline on the most recent commits again, you could just run `pachctl run pipeline cross-pipe`.

If you try to use a commit from a branch that is not an input branch, it will give you an error. 
Similarly, giving it multiple commits from the same branch will give an error.

Note that you do not need this (and shouldn't use it) to run a new pipeline you've just created. 
Pipelines will run automatically as you add new commits to their input branches.