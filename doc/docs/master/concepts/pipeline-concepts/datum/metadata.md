# Datum Metadata

Datum metadata provides the following information for any datums
processed by your pipelines:

- The amount of data that was uploaded and downloaded
- The time spend uploading and downloading data
- The total time spend processing
- Success/failure information, including any error encountered for failed datums
- The directory structure of input data that was seen by the job.

This information is
available through the `pachctl inspect datum` and `pachctl list datum`
commands or through their language client equivalents.


Once a pipeline has finished a job, **you can access metadata about the datums
processed during that job** in the associated meta [system repo](../../reference/system_repos.md).

!!! example

    ```shell
    pachctl list file edges.meta@master
    ```

    **System response:**

    ```shell
    NAME   TAG TYPE SIZE
    /meta/     dir  1.956KiB
    /pfs/      dir  371.9KiB
    ```

## Meta directory
The **meta directory holds each datum's JSON metadata**, and can be accessed using a `get file``:

!!! example

    ```shell
    pachctl get file edges.meta@master:/meta/002f991aa9db9f0c44a92a30dff8ab22e788f86cc851bec80d5a74e05ad12868/meta | jq
    ```

    **System response:**

    ```shell
    {
      "job": {
        "pipeline": {
          "name": "edges"
        },
        "id": "efca9595bdde4c0ba46a444a5877fdfe"
      },
      "inputs": [
        {
          "fileInfo": {
            ...
        }
      ],
      "hash": "28e6675faba53383ac84b899d853bb0781c6b13a90686758ce5b3644af28cb62f763",
      "stats": {
        "downloadTime": "0.103591200s",
        "processTime": "0.374824700s",
        "uploadTime": "0.001807800s",
        "downloadBytes": "80588",
        "uploadBytes": "38046"
      },
      "index": "1"
    }
    ```

## Pfs Directory
The **pfs directory has both the input and the output data that was committed in this datum**:

!!! example

    ```shell
    pachctl list file edges@stats:/002f991aa9db9f0c44a92a30dff8ab22e788f86cc851bec80d5a74e05ad12868/pfs
    ```

    **System response:**

    ```shell
    NAME                                                                         TYPE SIZE
    /002f991aa9db9f0c44a92a30dff8ab22e788f86cc851bec80d5a74e05ad12868/pfs/images dir  78.7KiB  
    /002f991aa9db9f0c44a92a30dff8ab22e788f86cc851bec80d5a74e05ad12868/pfs/out    dir  37.15KiB
    ```

