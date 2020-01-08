# Deferred Processing and Transactions

## Introduction
[Deferred processing](https://docs.pachyderm.com/latest/how-tos/deferred_processing/) 
is a Pachyderm technique for controlling when data gets processed.
Transactions are Pachyderm feature for assuring data from different repos get processed together.

This is an example which uses a simple DAG, 
based on our opencv example, 
to illustrate using deferred processing along with transactions.



## Pipelines

![Example DAG](./example_dag.png)

The DAG shown is a simple elaboration on the opencv example,
with pipeline and repo names chosen to avoid collisions with that example 
if installed in the same cluster.

The functionality is, of course, different.
The `edges_dp` pipeline will do edge detection on images committed to `images_dp_1`.
`montage_dp` will create a montage out of images committed to `images_dp_2` and images in the master branch of `edges_dp`. 
This is so it's easy to verify the files being processed
by visually inspecting the montage.

The most significant change from opencv in this example is the pipeline spec for `edges_dp`,
which has the `output_branch` attribute set to `dev`.

```json hl_lines="11"
{
    "pipeline": {
        "name": "edges_dp"
    },
    "input": {
        "pfs": {
            "glob": "/*",
            "repo": "images_dp_1"
        }
    },
    "output_branch": "dev",
    "transform": {
        "cmd": [ "python3", "/edges.py" ],
        "image": "pachyderm/opencv"
    }
}
```

This means that this pipeline, 
instead of putting its output in the `master` branch,
will put it in a branch called `dev`.

Since `montage_dp` is subscribed to the master branch of `edges_dp`, jobs will not trigger when the edges pipeline outputs files to the `dev` branch. Instead, to trigger a montage job, 
we can simply create a `master` branch attached to any commit in `edges_dp` to trigger the pipeline.

## Example run-through

### Deferred Processing

You should have a Pachyderm cluster set up 
and access to it configured from your local computer
prior to running this example.

1. Run the script `setup.sh` included in this repo.
   It will perform the following steps for you:
   
   ```sh
   pachctl create repo images_dp_1
   pachctl create repo images_dp_2
   pachctl create pipeline -f ./edges_dp.json 
   pachctl create pipeline -f ./montage_dp.json 
   pachctl put file images_dp_1@master -i ./images.txt
   pachctl put file images_dp_1@master -i ./images2.txt
   pachctl put file images_dp_2@master -i ./images3.txt
   pachctl create branch edges_dp@stable_1_0 --head dev^
   ```
   
2. Once the demo is loaded, 
   look at the commits in `edges_dp`.
   They should look something like this.
   
   ```sh
   $  pachctl list commit edges_dp
   REPO         BRANCH COMMIT                           FINISHED      SIZE PROGRESS DESCRIPTION
   edges_dp     dev    85ca689439b943978d273b311003d1bf 2 minutes ago 0B   -         
   edges_dp     dev    dc23aa4da6b94fefbebb6751c1c9b5a3 2 minutes ago 0B   -         
   ```
   
3. Remember that the `edges_dp` pipeline outputs to the `dev` branch.
   Since the `montage_dp` pipeline subscribes to the `master` branch,
   it won't be triggered when `edges_dp` jobs complete,
   since that output goes into the `dev` branch.

4. See that there are two commits in `edges_dp`,
   both in the `dev` branch that's the output for the pipeline.

   ```sh
   $ pachctl list commit edges_dp
   REPO         BRANCH COMMIT                           FINISHED      SIZE PROGRESS DESCRIPTION
   edges_dp     dev    85ca689439b943978d273b311003d1bf 2 minutes ago 0B   -         
   edges_dp     dev    dc23aa4da6b94fefbebb6751c1c9b5a3 2 minutes ago 0B   -         
   ```
    
5. We've also created a stable branch that points at that first commit,
   with the `dev` branch at the most recent commit.
   The `master` branch, of course, has no commits.
   Take note of the commit ids and match them to the ids above.
   
   ```sh
   $ pachctl list branch edges_dp
   BRANCH     HEAD                             
   stable_1_0 dc23aa4da6b94fefbebb6751c1c9b5a3 
   master     -                                
   dev        85ca689439b943978d273b311003d1bf 
   ```

6. If we look at the jobs for this DAG, 
   we'll see that there are three jobs:
   - one 0-datum job for `montage_dp` and 
   - two jobs for `edges_dp` with the appropriate number of datums in each.
   
   This is what we'd expect.
   There's no data in the master branch of `edges_dp`, 
   so an empty job was created in `montage_dp`
   when data was commited to `images_dp_2`
   because of its `cross` input.
   
   ```sh
   $ pachctl list job
   ID                               PIPELINE           STARTED        DURATION  RESTART PROGRESS  DL       UL STATE   
   7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         12 minutes ago 1 second  0       0 + 0 / 0 0B       0B success 
   f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           13 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 0B success 
   f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           13 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 0B success 
   ```
   
7. Now we commit a file to `images_dp_1`. 

   ```sh
   $ pachctl put file images_dp_1@master:1VqcWw9.jpg -f http://imgur.com/1VqcWw9.jpg
   ```

8. We can see that one job was triggered in `edges_dp`
   with the one datum we committed, above, processed
   and the three existing datums skipped.
   We can also confirm that no job was triggered for `montage_dp`. 

   ```sh
   $ pachctl list job
   ID                               PIPELINE           STARTED        DURATION  RESTART PROGRESS  DL       UL STATE   
   c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           11 seconds ago 3 seconds 0       1 + 3 / 4 175.1KiB 0B success 
   7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         20 minutes ago 1 second  0       0 + 0 / 0 0B       0B success 
   f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           20 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 0B success 
   f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           20 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 0B success 
   ```

9. Now we want to trigger `montage_dp` on the stable set of data in our `stable_1_0` branch.
   We can accomplish that by creating a `master` branch with `stable_1_0` as its head.
   
   ```sh
   $ pachctl create branch edges_dp@master --head stable_1_0
   ```
   
10. Listing jobs will show that a job got triggered on `montage_dp`.

    ```sh
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE   
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         31 seconds ago     4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           About a minute ago 3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         21 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           21 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           21 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

11. If we want to trigger another job in `montage_dp`,
    based on what's currently in the `dev` branch,
    we simply update `master` to point at `dev`.
    
    ```sh
    $ pachctl create branch edges_dp@master --head dev
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         About a minute ago 3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         3 minutes ago      4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           3 minutes ago      3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         24 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           24 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           24 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

12. We won't cause another job to trigger on `montage_dp` if we commit more data to `images_dp_1`.
    It will only trigger a job in `edges_dp`.

    ```sh
    $ pachctl put file images_dp_1@master:2GI70mb.jpg -f http://imgur.com/2GI70mb.jpg
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE   
    96439afdaf8249be81773c89fc87b366 edges_dp           31 seconds ago     3 seconds 0       1 + 4 / 5 204KiB   0B       success 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         About a minute ago 3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         3 minutes ago      4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           3 minutes ago      3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         24 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           24 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           24 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

13. We can create another "stable" branch called `stable_1_1`.
    
    ```sh
    $ pachctl create branch edges_dp@stable_1_1 --head master
    ```

14. We can move master to our old stable branch, `stable_1_0`,
    which will create a job in `montage_dp` which skips its datum,
    as that data was already processed.
    
    ```sh
    $ pachctl create branch edges_dp@master --head stable_1_0
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE   
    01bb8e400c054a18a5c2c691f0b5a65d montage_dp         11 seconds ago     2 seconds 0       0 + 1 / 1 0B       0B       success 
    96439afdaf8249be81773c89fc87b366 edges_dp           42 seconds ago     3 seconds 0       1 + 4 / 5 204KiB   0B       success 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         About a minute ago 3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         3 minutes ago      4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           3 minutes ago      3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         24 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           24 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           24 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

### Transactions

15. If you want to run a particular set of data in `images_dp_2` 
    against a particular branch of `edges_dp`,
    you need to perform two operations
    - commit data to `images_dp_2` and
    - point `edges_dp@master` to the specific commit of interest.
    
    If you do not use a transaction, this will result in two jobs being triggered, one for the new commit and a second when we move `edges_dp@master` branch.
    - `images_dp_2@master` running against whatever is currently in `edges_dp@master`
    - `images_dp_2@master` running against whatever you set `edges_dp@master` to
    
    Remember that in step 14 above, 
    we performed the `create branch` operation.
    Now we perform the commit
    and see that another job got triggered.
    
    ```sh
    $ pachctl put file images_dp_2@master:3Kr6Mr6.jpg  -f http://imgur.com/3Kr6Mr6.jpg
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE   
    1c8d2bb48bca451ba43fdc5308a24d6d montage_dp         9 seconds ago      4 seconds 0       1 + 0 / 1 770.7KiB 906.7KiB success 
    01bb8e400c054a18a5c2c691f0b5a65d montage_dp         About a minute ago 2 seconds 0       0 + 1 / 1 0B       0B       success 
    96439afdaf8249be81773c89fc87b366 edges_dp           About a minute ago 3 seconds 0       1 + 4 / 5 204KiB   0B       success 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         3 minutes ago      3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         4 minutes ago      4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           4 minutes ago      3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         25 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           25 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           25 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```
    
16. If you want to just have one job 
    where `images_dp_2@master` runs against whatever you set `edges_dp@master` to,
    you can use Pachyderm transactions.
    First step is to start a transaction.
    
    ```sh
    $ pachctl start transaction
    Started new transaction: 018b9e52-e483-4a0a-accd-cef2656f7b8c
    ```
    
17. Once the transaction is started,
    you start all commits and branch creations 
    within the scope of the transaction. 
    
    ```sh
    $ pachctl start commit  images_dp_2@master
    Added to transaction: 018b9e52-e483-4a0a-accd-cef2656f7b8c
    e6e68bd54b9049658fca31aa7e905fcd
    $ pachctl create branch edges_dp@master --head stable_1_1
    Added to transaction: 018b9e52-e483-4a0a-accd-cef2656f7b8c
    ```


18. Before you put any files in an repos, 
    you need to finish the transaction.
    That will group all the commits and branches together,
    triggering when the last commit in the transaction is finished.
    
    ```sh
    $ pachctl finish transaction
    Completed transaction with 2 requests: 018b9e52-e483-4a0a-accd-cef2656f7b8c
    ```
    
19. No job has yet been triggered.
    We'll commit a file, 
    and job list will show no new jobs.
    
    ```sh
    $ pachctl put file images_dp_2@master:9iIlokw.jpg -f http://imgur.com/9iIlokw.jpg
    $ pachctl list job
    ID                               PIPELINE           STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE   
    1c8d2bb48bca451ba43fdc5308a24d6d montage_dp         About a minute ago 4 seconds 0       1 + 0 / 1 770.7KiB 906.7KiB success 
    01bb8e400c054a18a5c2c691f0b5a65d montage_dp         About a minute ago 2 seconds 0       0 + 1 / 1 0B       0B       success 
    96439afdaf8249be81773c89fc87b366 edges_dp           About a minute ago 3 seconds 0       1 + 4 / 5 204KiB   0B       success 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         3 minutes ago      3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         4 minutes ago      4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           4 minutes ago      3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         25 minutes ago     1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           25 minutes ago     4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           25 minutes ago     3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

20.   Finishing the commit that we started during the transaction finally starts the job.

    ```
    $ pachctl finish commit images_dp_2@master
    $ pachctl list job
    ID                               PIPELINE           STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE   
    ce8c42f40d4a437598c128f768ed2349 montage_dp         14 seconds ago 5 seconds 0       1 + 0 / 1 958.1KiB 1.18MiB  success 
    1c8d2bb48bca451ba43fdc5308a24d6d montage_dp         3 minutes ago  4 seconds 0       1 + 0 / 1 770.7KiB 906.7KiB success 
    01bb8e400c054a18a5c2c691f0b5a65d montage_dp         4 minutes ago  2 seconds 0       0 + 1 / 1 0B       0B       success 
    96439afdaf8249be81773c89fc87b366 edges_dp           5 minutes ago  3 seconds 0       1 + 4 / 5 204KiB   0B       success 
    4b05f2c0ddf14ab0a4ca8f4ee2d6d9ff montage_dp         6 minutes ago  3 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    1a8d2fefc2ba450e9102ffd6b74e503d montage_dp         7 minutes ago  4 seconds 0       1 + 0 / 1 693.8KiB 673.6KiB success 
    c5feebe5c6ae48e0adc4f5bd7be9a9ac edges_dp           8 minutes ago  3 seconds 0       1 + 3 / 4 175.1KiB 0B       success 
    7d6aa9bbe1484e2e9ac60a7f84147e1b montage_dp         28 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success 
    f7a7d71dd7e74a3f9c48d4b9d2897ba3 edges_dp           28 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 0B       success 
    f9c620d1a4cc421a9c3923b749e9e6ed edges_dp           28 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 0B       success 
    ```

## Summary

Deferred processing with transactions in Pachyderm 
will give you fine-grained control of jobs and datums
while preserving Pachyderm's advantages of data lineage and incremental processing.
