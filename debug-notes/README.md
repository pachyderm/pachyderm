2022-05-05
----------
- Add quotes to ETag in object.go
  - *Hypothesis*: I know from [this bug](...) that golang's implementation of
    ServeContent applies the "must have quotes" etag requirement. I also know
    that s2/object.go calls ServeContent, so I hypothesize that that's the call
    that's rejecting Spark's S3 requests. A benefit of this being the case is
    that the HTTP request handler calls ServeContent(req, …) so I could just
    edit the ETag request and add quotes before calling ServeContent
  - *Experiment* Edit the ServeContent call site in object.go. The hypothesis
    predicts that our Spark example will start passing
  - *Result* object.go already adds etag quotes! And yet the example doesn't
    pass, so hypothesis disproven.
  - *Notes* Why doesn't it work?

2022-03-22
----------
- Trying to debug the S3 gateway etag/quotes issue that albert found
  - working in code/20220321-s3g_get_spark_working
- Alon seems to have broken port-forward in 7357; sent him a message
  - For now, I can work around the issue by running `pachctl port-forward --remote-port=1650 --remote-s3gateway-port=1600 --remote-oidc-port=1657 --remote-dex-port=1658`
  - I can log in using `pachctl auth login` and then username=admin and password=password. I reverse-engineered this with:
    ```
    $ kc get secret/pachyderm-bootstrap-config -o json | jq -r .
    data.idps | base64 -d

    - id: test
      name: test
      type: mockPassword
      jsonConfig: '{"username": "admin", "password": "password"}'
    ```
    -	It took me a v. long time to realize that:
      - a) the config pod activates auth by default if passed an enterprise token (which my `restart_minikube` script now does for me)
      - b) when this happens, auth is configured using kubernetes secrets deployed by helm (etc/helm/pachyderm/templates/pachd/config-secret.yaml), which the config-pod mounts
      - c) Even though nothing relevant is printed and my newly-deployed cluster appears as an impenetrable box, after deploying auth, I can get the info I need to log in by inspecting this kubernetes secret post-facto
  - I realized I can also use the root token with:
    ```
       $ kc get secret/pachyderm-bootstrap-config -o json | jq -r .data.rootToken | base64 -d
       CYTXz...
    ```
  - To load the data my little spark job expects, I can run: `pachctl put file fhir@subset-10:/binaryFile -f mascot.txt`
  - Then, the spark job _seems_ to succeed, but it actually doesn't; spark dataframes are lazy, and to force a read you have to call `df.show()` or some such on the result of the dataframe constructor. Then you get the etag error!
    - I just did `print(df.show())`

2022-03-21
----------
Modified instructions from Albert:
Requirements:
create a Python virtualenv, and activate it:
```
mkdir venv && pushd venv && python3 -m venv . && popd
source ./venv/bin/activate

python3 -m pip install pyspark

curl -L \
  -o "${VIRTUAL_ENV}/lib/python3.9/site-packages/pyspark/jars/hadoop-aws-3.3.1.jar" \
  https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.1/hadoop-aws-3.3.1.jar
curl -LO 
curl -L \
  -o "${VIRTUAL_ENV}/lib/python3.9/site-packages/pyspark/jars/aws-java-sdk-bundle-1.12.153.jar" \
  https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.153/aws-java-sdk-bundle-1.12.153.jar
```

Set up pachyderm repo and data
```
pachctl create repo fhir
pachctl put file fhir@subset-10:/sample.json << EOF 
{"name": "Gilbert", "wins": [["straight", "7♣"], ["one pair", "10♥"]]}
{"name": "Alexa", "wins": [["two pair", "4♠"], ["two pair", "9♠"]]}
{"name": "May", "wins": []}
{"name": "Deloise", "wins": [["three of a kind", "5♣"]]}
EOF
```

Run the following Python script
```
from pyspark.sql import SparkSession


spark = SparkSession.builder \
            .master("local[1]") \
            .appName("my_app") \
            .getOrCreate()

spark._jsc.hadoopConfiguration().set("fs.s3a.endpoint", "http://localhost:30600")
spark._jsc.hadoopConfiguration().set("fs.s3a.access.key", "foo")
spark._jsc.hadoopConfiguration().set("fs.s3a.secret.key", "foo")
spark._jsc.hadoopConfiguration().set("fs.s3a.impl","org.apache.hadoop.fs.s3a.S3AFileSystem")
spark._jsc.hadoopConfiguration().set("fs.s3a.connection.ssl.enabled", "false")
spark._jsc.hadoopConfiguration().set("fs.s3a.metadatastore.impl", "org.apache.hadoop.fs.s3a.s3guard.NullMetadataStore")
spark._jsc.hadoopConfiguration().set("fs.s3a.path.style.access", "true")

spark.sparkContext.setLogLevel("debug")


df = spark.read.format("json").load('s3a://subset-10.fhir/sample.json')
df.show()
```

To actually run the Spark script, you need to reference the downloaded Jars
```
spark-submit --jars aws-java-sdk-bundle-1.12.153.jar,hadoop-aws-3.3.1.jar test-spark-pach.py
```

#### Observed errors
Type 1 error: `HTTP/1.1 412 Precondition Failed`
Example
```
>> GET http://localhost:30600 /subset-10.fhir/sample.json Headers: (amz-sdk-invocation-id: 9fe61a83-5486-4bea-722e-005c8d2b029e, Content-Type: application/octet-stream, If-Match: 3083f182ca306d85c0e0fa2b7e70d844cd2cbceb772ccc9543f55989fd9c183a ...)

<< HTTP/1.1 412 Precondition Failed
```
This is more straight forward to explain. It seems like Spark is making a bad request here, and that the Etag in the If-Match should be quoted.

```
curl -i -H 'If-Match: 3083f182ca306d85c0e0fa2b7e70d844cd2cbceb772ccc9543f55989fd9c183a' -X GET  "localhost:30600/subset-10.fhir/sample.json" 
HTTP/1.1 412 Precondition Failed
```
vs
```
curl -i -H 'If-Match: "3083f182ca306d85c0e0fa2b7e70d844cd2cbceb772ccc9543f55989fd9c183a"' -X GET  "localhost:30600/subset-10.fhir/sample.json"
HTTP/1.1 200 OK
```

Type 2 error: `Change reported by S3`
```
22/03/10 15:09:09 ERROR Executor: Exception in task 0.0 in stage 0.0 (TID 0)
org.apache.hadoop.fs.s3a.RemoteFileChangedException: open 's3a://subset-10.fhir/sample.json': Change reported by S3 during open at position 0. ETag 3083f182ca306d85c0e0fa2b7e70d844cd2cbceb772ccc9543f55989fd9c183a was unavailable
```
Tried to disable S3Guard on client side by setting NullMetadataStore, but to no avail
```
22/03/10 15:09:07 DEBUG S3Guard: Using NullMetadataStore metadata store for s3a filesystem
22/03/10 15:09:07 DEBUG S3AFileSystem: S3Guard is disabled on this bucket: subset-10.fhir
```

Hopefully by fixing the 412 error, we can also fix the S3Guard error.

2022-03-05
----------
Steps to get this thing working:
- Testing this by running `spark-submit ./spark-pach-example.py`. I think I got an error message when I ran `spark-pach-example.py` directly, that instructed me to do this
- python3 -m venv .
- source ./bin/activate
- python3 -m pip install pyspark (source: guessing that this might work?)
- Fix up python:
  - Add imports
    - `NameError: name 'SparkConf' is not defined`
      - Add `import SparkConf from pyspark`, per https://stackoverflow.com/questions/62483215/how-to-resolve-the-error-nameerror-name-sparkconf-is-not-defined-in-pycharm
    - `ImportError: cannot import name 'SparkSession' from 'pyspark' (/home/m/code/20210804-pach-1.13.x/2022-02-07_anthem_debug/lib/python3.9/`
      - Add `import SparkConf from pyspark.sql` per https://stackoverflow.com/questions/40838040/cannot-import-name-sparksession/40838356
- Error: `Caused by: java.net.UnknownHostException: kubernetes.default.svc.cluster.local: Name or service not known`
  - Change spark address to `https://192.168.39.149:8443` per my kubeconfig
- Error: hanging with repeated log lines like:
  ```
     22/02/07 21:06:12 INFO ExecutorPodsAllocator: Going to request 2 executors from Kubernetes for ResourceProfile Id: 0, target: 2, known: 0, sharedSlotFromPendingPods: 2147483647.
     22/02/07 21:06:12 WARN ExecutorPodsSnapshotsStoreImpl: Exception when notifying snapshot subscriber.
     org.apache.spark.SparkException: Must specify the executor container image
  ```
  - Solution: change spark address to 'local' so that it runs locally in a single thread, per https://stackoverflow.com/questions/32356143/what-does-setmaster-local-mean-in-spark
- Error: `Caused by: java.lang.ClassNotFoundException: Class org.apache.hadoop.fs.s3a.S3AFileSystem not found`
  - Resource: I need `aws-java-sdk` per https://stackoverflow.com/questions/58415928/spark-s3-error-java-lang-classnotfoundexception-class-org-apache-hadoop-f
  - Docs: https://hadoop.apache.org/docs/r3.1.2/hadoop-aws/tools/hadoop-aws/troubleshooting_s3a.html#ClassNotFoundException:_org.apache.hadoop.fs.s3a.S3AFileSystem
  - I figured out where `$SPARK_HOME` was by adding `print(">>>> "+os.environ["SPARK_HOME"])` to the spark script, and found that it was in `/home/m/code/20210804-pach-1.13.x/2022-02-07_anthem_debug/lib/python3.9/site-packages/pyspark`
  - `cd lib/python3.9/site-packages/pyspark/jars`
  - `curl -LO https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.1/hadoop-aws-3.3.1.jar` (found this 
  - `curl -LO https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.153/aws-java-sdk-bundle-1.12.153.jar`
- Error: `Caused by: com.amazonaws.SdkClientException: Unable to load AWS credentials from environment variables (AWS_ACCESS_KEY_ID (or AWS_ACCESS_KEY) and AWS_SECRET_KEY (or AWS_SECRET_ACCESS_KEY))`
  -
