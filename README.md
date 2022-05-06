# Debugging S3 + Spark

This branch is for debugging the currently-broken interaction between Pachyderm and Spark:
- See debug-notes for a repro, a script for testing, and previous notes
- Instead of depending on github.com/pachyderm/s2, this branch has been hacked to depend on a clone of that repo in the top-level s2 folder. This makes iterating on pach and s2 in tandem easier. At the end, if changes to s2 are needed, `diff` can generate a patch to the s2 code that can be applied in the s2 repo and released there officially.
