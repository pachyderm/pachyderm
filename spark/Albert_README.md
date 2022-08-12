### Requirements
- A Pachyderm repo to write to
- `pip install pyspark (the user is using 3.1.2)`
- Java 11
  - `sudo apt install openjdk-11-jdk openjdk-11-jre`
  - `export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64`
- Jars: 
  - `wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.3/hadoop-aws-3.3.3.jar`
  - `wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.264/aws-java-sdk-bundle-1.12.264.jar`
- test.py
    ```
    from pyspark.sql import SparkSession, Row, DataFrame
    from pyspark.context import SparkContext
    from pyspark import SparkConf
    import time
    import os

    conf = SparkConf()
    conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:30600")
    conf.set('spark.hadoop.fs.s3a.impl', "org.apache.hadoop.fs.s3a.S3AFileSystem")
    conf.set('spark.hadoop.fs.s3a.access.key', 'lemon')
    conf.set('spark.hadoop.fs.s3a.secret.key', 'lemon')
    conf.set('spark.hadoop.fs.s3a.path.style.access', 'true')
    conf.set('spark.hadoop.fs.s3a.connection.ssl.enabled', 'false')
    sc = SparkContext(conf=conf)
    sc.setLogLevel("DEBUG")

    # confirm config is applied to this session
    spark = SparkSession.builder.getOrCreate()
    conf = spark.sparkContext.getConf()
    sc = spark.sparkContext
    conf = sc.getConf()
    print(sc.getConf().getAll())

    # create some silly data
    df = spark.createDataFrame([ Row(a=1, b=2.,) ])
    df.show()

    df.write.parquet('s3a://master.rando2/o', mode="overwrite")
    ```

### Steps
- Create bucket: `pachctl create repo rando2; pachctl create branch rando2@master`
- `pachctl port-forward` to expose S3G
- run Spark:
    ```
    spark-submit --jars 'hadoop-aws-3.3.3.jar,aws-java-sdk-bundle-1.12.264.jar' test.py

    Error:
    Traceback (most recent call last):
      File "/app/main.py", line 40, in <module>
          df.write.parquet('s3a://out/a', mode="overwrite")
            File "/usr/local/spark/python/lib/pyspark.zip/pyspark/sql/readwriter.py", line 1140, in parquet
              File "/usr/local/spark/python/lib/py4j-0.10.9.5-src.zip/py4j/java_gateway.py", line 1321, in __call__
                File "/usr/local/spark/python/lib/pyspark.zip/pyspark/sql/utils.py", line 190, in deco
                  File "/usr/local/spark/python/lib/py4j-0.10.9.5-src.zip/py4j/protocol.py", line 326, in get_return_value
                  py4j.protocol.Py4JJavaError: An error occurred while calling o223.parquet.
                  : org.apache.hadoop.fs.s3a.AWSBadRequestException: PUT 0-byte object  on a/_temporary/0: com.amazonaws.services.s3.model.AmazonS3Exception: Invalid file path (Service: Amazon S3; Status Code: 400; Error Code: InvalidFilePath; Request ID: c8ed08bc-9da4-4632-bf7a-3941732eef7b; S3 Extended Request ID: null; Proxy: null), S3 Extended Request ID: null:InvalidFilePath: Invalid file path (Service: Amazon S3; Status Code: 400; Error Code: InvalidFilePath; Request ID: c8ed08bc-9da4-4632-bf7a-3941732eef7b; S3 Extended Request ID: null; Proxy: null)
    ```
- a closer look:
    ```
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » PUT /out/a/_temporary/0/ HTTP/1.1
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Host: s3-5853903aba486d3f11b00f44d3e2a915.default:1600
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » amz-sdk-invocation-id: 2c067119-ba7f-405e-3e77-0d62c62bd391
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » amz-sdk-request: attempt=1;max=21
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » amz-sdk-retry: 0/0/500
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Authorization: AWS4-HMAC-SHA256 Credential=lemon/20220722/5853903aba486d3f11b00f44d3e2a915/s3/aws4_request, SignedHeaders=amz-sdk-invocation-id;amz-sdk-request;amz-sdk-retry;content-length;content-type;host;user-agent;x-amz-content-sha256;x-amz-date;x-amz-decoded-content-length, Signature=09ac6419974746191945abcf39fa3a59f7658c5fbe7750be5079417926845d1a
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Content-Type: application/x-directory
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » User-Agent: Hadoop 3.3.2, aws-sdk-java/1.12.264 Linux/5.10.102.1-microsoft-standard-WSL2 OpenJDK_64-Bit_Server_VM/17.0.3+7-Ubuntu-0ubuntu0.20.04.1 java/17.0.3 scala/2.13.8 vendor/Private_Build cfg/retry-mode/legacy
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » x-amz-content-sha256: STREAMING-AWS4-HMAC-SHA256-PAYLOAD
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » X-Amz-Date: 20220722T164544Z
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » x-amz-decoded-content-length: 0
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Content-Length: 86
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Connection: Keep-Alive
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 » Expect: 100-continue
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "PUT /out/a/_temporary/0/ HTTP/1.1[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Host: s3-5853903aba486d3f11b00f44d3e2a915.default:1600[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "amz-sdk-invocation-id: 2c067119-ba7f-405e-3e77-0d62c62bd391[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "amz-sdk-request: attempt=1;max=21[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "amz-sdk-retry: 0/0/500[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Authorization: AWS4-HMAC-SHA256 Credential=lemon/20220722/5853903aba486d3f11b00f44d3e2a915/s3/aws4_request, SignedHeaders=amz-sdk-invocation-id;amz-sdk-request;amz-sdk-retry;content-length;content-type;host;user-agent;x-amz-content-sha256;x-amz-date;x-amz-decoded-content-length, Signature=09ac6419974746191945abcf39fa3a59f7658c5fbe7750be5079417926845d1a[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Content-Type: application/x-directory[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "User-Agent: Hadoop 3.3.2, aws-sdk-java/1.12.264 Linux/5.10.102.1-microsoft-standard-WSL2 OpenJDK_64-Bit_Server_VM/17.0.3+7-Ubuntu-0ubuntu0.20.04.1 java/17.0.3 scala/2.13.8 vendor/Private_Build cfg/retry-mode/legacy[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "x-amz-content-sha256: STREAMING-AWS4-HMAC-SHA256-PAYLOAD[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "X-Amz-Date: 20220722T164544Z[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "x-amz-decoded-content-length: 0[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Content-Length: 86[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Connection: Keep-Alive[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "Expect: 100-continue[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "HTTP/1.1 100 Continue[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "[\r][\n]"
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « HTTP/1.1 100 Continue
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "0;chunk-signature=2b948fbd316d2b1111bfe5f4d46b5c0c9efd7f0ce0c288500d1bf9079eba79b5[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 » "[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "HTTP/1.1 400 Bad Request[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "Content-Type: application/xml[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "X-Amz-Id-2: c8ed08bc-9da4-4632-bf7a-3941732eef7b[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "X-Amz-Request-Id: c8ed08bc-9da4-4632-bf7a-3941732eef7b[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "Date: Fri, 22 Jul 2022 16:45:44 GMT[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "Content-Length: 218[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "[\r][\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "<?xml version="1.0" encoding="UTF-8"?>[\n]"
    22/07/22 16:45:44 DEBUG wire: http-outgoing-0 « "<Error><Code>InvalidFilePath</Code><Message>Invalid file path</Message><Resource>/out/a/_temporary/0/</Resource><RequestId>c8ed08bc-9da4-4632-bf7a-3941732eef7b</RequestId></Error>"
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « HTTP/1.1 400 Bad Request
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « Content-Type: application/xml
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « X-Amz-Id-2: c8ed08bc-9da4-4632-bf7a-3941732eef7b
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « X-Amz-Request-Id: c8ed08bc-9da4-4632-bf7a-3941732eef7b
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « Date: Fri, 22 Jul 2022 16:45:44 GMT
    22/07/22 16:45:44 DEBUG headers: http-outgoing-0 « Content-Length: 218
    ```
- This is due to https://github.com/pachyderm/pachyderm/blob/master/src/server/pfs/s3/object.go#L126

### Discussion

What’s the root cause? PFS doesn’t allow objects with trailing / but Spark is trying to create them as empty directories to hold temporary files.

#### How do we fix this?
##### Option 1:  trim the / suffix in all requests
Trying this right now, but getting some other errors:
```
py4j.protocol.Py4JJavaError: An error occurred while calling o223.parquet.
: org.apache.hadoop.fs.s3a.AWSClientIOException: PUT 0-byte object  on o/_temporary/0: com.amazonaws.SdkClientException: Unable to verify integrity of data upload. Client calculated content hash (contentMD5: 1B2M2Y8AsgTpgAmY7PhCfg== in base 64) didn't match hash (etag: 7c09f7c4d76ace86e1a7e1c7dc0a0c7edcaa8b284949320081131976a87760c3 in hex) calculated by Amazon S3.  You may need to delete the data stored in Amazon S3. (metadata.contentMD5: null, md5DigestStream: com.amazonaws.services.s3.internal.MD5DigestCalculatingInputStream@70cb7a34, bucketName: master.rando, key: o/_temporary/0/): Unable to verify integrity of data upload. Client calculated content hash (contentMD5: 1B2M2Y8AsgTpgAmY7PhCfg== in base 64) didn't match hash (etag: 7c09f7c4d76ace86e1a7e1c7dc0a0c7edcaa8b284949320081131976a87760c3 in hex) calculated by Amazon S3.  You may need to delete the data stored in Amazon S3. (metadata.contentMD5: null, md5DigestStream: com.amazonaws.services.s3.internal.MD5DigestCalculatingInputStream@70cb7a34, bucketName: master.rando, key: o/_temporary/0/)
```
**Update**: I think the issue is something to do with this https://hadoop.apache.org/docs/stable/hadoop-aws/tools/hadoop-aws/troubleshooting_s3a.html#:~:text=Other%20Errors-,SdkClientException%20Unable%20to%20verify%20integrity%20of%20data%20upload,-Something%20has%20happened
also here are a list of the different signing algorithms: https://docs.cloudera.com/HDPDocuments/HDP3/HDP-3.0.1/bk_cloud-data-access/content/s3-third-party.html
Basically the client is hashing the data as it’s uploading, then verifies against the response’s etag once its done. However, our etag is based on PFS’s mechanism of hashing, which is probably different from how S3 does it.

So what we should do is, first confirm what algorithm the client uses to hash (most likely Md5), and also how does it apply MD5? This is most likely also documented by AWS S3.

Then we simply update our S3G’s code to apply the same hashing mechanism. https://github.com/pachyderm/pachyderm/tree/albert/CORE-801/ignore-trailing-slash

##### Option 2: make PFS work with trailing / by allowing an empty directory.
But this was an explicit design decision for performance reasons

Short term goal: get S3G to work with Spark writes.

Long term goal: more robust Spark integration.

Spark is too popular for us to ignore. We need some kind of integration with Spark. S3G is probably not the long term solution since there’s performance bottlenecks for writes.

Ideas for a more robust Spark integration:
Implement our own Java client for Spark that converts Spark data operations to PFS operations, bypassing S3G all together.
Allow Spark to talk to Object storage directly, then have some kind of background sync that writes that data back into PFS
