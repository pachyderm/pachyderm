# Create a Join Pipeline

In this example, we will create a join pipeline.
A join pipeline executes your code on files that match
a specific naming pattern. For example, you have two
repositories — one that stores readings from temperature
sensors and the other that stores parameters for these
sensors. You want to get the sum of all lines in the
files in the readings `repository` and multiply them
to the sum of lines in matching files in the
parameters repository.

The repositories have the following structures:

* `readings`:

    ```bash
    ├── ID1234
        ├── file1.txt
        ├── file2.txt
        ├── file3.txt
        ├── file4.txt
        ├── file5.txt
    ```

 * `parameters`:

    ```bash
    ├── file1.txt
    ├── file2.txt
    ├── file3.txt
    ├── file4.txt
    ├── file5.txt
    ```

## Prerequisites

You must have the following installed on your computer
to complete this example:

* Minikube or Docker Desktop
* `pachctl` 1.9.5 or later
* `pachd` 1.9.5 or later


## Step 1 - Prepare your Environment

`Makefile` in this directory creates everything you need for
you to test this example. The `Makefile` targets create all the
needed files, build and execute a Docker container
that runs the code in [joins.py]().

To set up your environment, complete the following steps:

1. From the `pachyderm/examples/joins/` directory, run:

   ```bash
   make setup
   ```

   This command performs the following steps:

   * Creates the `readings/ID1234` and  parameters` directories with the
   structures described above.

   * Builds a Docker container named joins-example with the `joins.py` code
   added to it.

1. Start the `joins-example` container and mount the `pfs/` directory:

   ```bash
   make run
   ```

## Step 2 - Set up the Pachyderm Repositories

You need to configure Pachyderm the `readings` and `parameters` repositories
in Pachyderm and upload the files in them.

To set up the Pachyderm repositories, complete the following steps:

1. Create the repositories:

   ```bash
   $ pachctl create repo readings
   $ pachctl create repo parameters
   ```

1. Upload the test files to the `readings` repository:

   ```bash
   $ pachctl put file -r readings1@master -f ID1234
   ```

1. Upload the test files to the `parameters` repository:

   ```bash
   $ pachctl put file -r params1@master:/ -f parameters
   ```

## Step 3 - Run the Pipeline

When your repositories are ready, create and run the pipeline
from the [joins.json]() file by completing the following steps:

1. Run:

   ```bash
   $ pachctl create pipeline -f joins.json
   ```

1. View the job status by running:

   ```bash
   $ pachctl list job
   ID                               PIPELINE STARTED       DURATION           RESTART PROGRESS  DL   UL  STATE
   7390b0c29ac247c893422a5c04565719 joins    2 seconds ago Less than a second 0       6 + 0 / 6 108B 24B success
   ```

   In the output above, you can see that Pachyderm processes six datums.

1. Get the contents of the file to view the result:

   ```bash
   $ pachctl list file joins@master
   NAME       TYPE SIZE
   /file1.txt file 4B
   /file2.txt file 4B
   /file3.txt file 4B
   /file4.txt file 4B
   /file5.txt file 4B
   /file6.txt file 4B
   ```

1. Get the contents of the file to view the result:

   ```
   $ pachctl get file joins@master:/file1.txt
   225
   ```
