# Writing to Pachyderm from Spark using the S3 Gateway

In Pachyderm 2.3.2+, Pachyderm's S3 Gateway has been made compatible with Spark's s3a adapter.
This means that you can write to a Pachyderm repo from a distributed Spark cluster.

You can connect Spark to Pachyderm's S3 gateway with, for example, options like the following:

```python
conf = SparkConf()
conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:30600")
conf.set('spark.hadoop.fs.s3a.impl', "org.apache.hadoop.fs.s3a.S3AFileSystem")
conf.set('spark.hadoop.fs.s3a.access.key', 'anything_matching')
conf.set('spark.hadoop.fs.s3a.secret.key', 'anything_matching')
conf.set('spark.hadoop.fs.s3a.path.style.access', 'true')
conf.set('spark.hadoop.fs.s3a.connection.ssl.enabled', 'false')
```

You may need to customize these for your own Pachyderm installation.

You need to ensure you open a commit for the duration of the Spark job, otherwise you will see errors.
This is because Spark's s3a driver writes temporary files with the same names as directories. This works in S3, but not in Pachyderm when closing commits: in this case you get a "file / directory path collision" error on the commits that the S3 gateway generates. Holding a commit open throughout the Spark job allows it to complete without errors.

You can hold a commit open during a job with `python_pachyderm`.

```python
with client.commit("spark_s3g_demo", "master") as commit:
    print(f"Opening commit {commit} for spark job")
    df.coalesce(1).write.format("parquet").mode("overwrite").save("s3a://master.spark_s3g_demo/example")
    print(f"Closing {commit}")
```

Another issue is that the Java S3 driver assumes that ETags are generated as MD5 hashes. Pachyderm doesn't do that, so we have to disable integrity checking on writes.

You can do that with:
```python
sc = SparkContext(conf=conf)
sc.setLogLevel("DEBUG")
sc.setSystemProperty("com.amazonaws.services.s3.disablePutObjectMD5Validation", "true")
```

# Complete example

Putting this all together, we get the following example of writing to Pachyderm from pyspark.
Start by creating the repo and the branch:
```
pachctl create repo spark_s3g_demo
pachctl create branch spark_s3g_demo@master
```

Install the dependencies, Java:
```
sudo apt install openjdk-11-jdk openjdk-11-jre
export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
```
PySpark:
```
pip3 install pyspark==3.3.0
```
Some jars:
```
wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.3/hadoop-aws-3.3.3.jar
wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.264/aws-java-sdk-bundle-1.12.264.jar
```

Now write this `spark.py`:

```python
from pyspark.sql import SparkSession, Row, DataFrame
from pyspark.context import SparkContext
from pyspark import SparkConf
import time
import os

conf = SparkConf()
conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:30600")
conf.set('spark.hadoop.fs.s3a.impl', "org.apache.hadoop.fs.s3a.S3AFileSystem")
conf.set('spark.hadoop.fs.s3a.access.key', 'anything_matching')
conf.set('spark.hadoop.fs.s3a.secret.key', 'anything_matching')
conf.set('spark.hadoop.fs.s3a.path.style.access', 'true')
conf.set('spark.hadoop.fs.s3a.connection.ssl.enabled', 'false')

sc = SparkContext(conf=conf)
sc.setLogLevel("DEBUG")
sc.setSystemProperty("com.amazonaws.services.s3.disablePutObjectMD5Validation", "true")

# confirm config is applied to this session
spark = SparkSession.builder.getOrCreate()
conf = spark.sparkContext.getConf()
sc = spark.sparkContext
conf = sc.getConf()
print(sc.getConf().getAll())

# create some example data
df = spark.createDataFrame([ Row(a=1, b=2.,) ])
df.show()

repo = "spark_s3g_demo"
branch = "master"

import python_pachyderm
client = python_pachyderm.Client()

with client.commit(repo, branch) as commit:
    print(f"Opening commit {commit} for spark job")
    df.coalesce(1).write.format("parquet").mode("overwrite").save(f"s3a://{branch}.{repo}/example_data")
    print(f"Closing {commit}")
```

And run it with:
```
spark-submit --jars 'hadoop-aws-3.3.3.jar,aws-java-sdk-bundle-1.12.264.jar' spark.py 2>&1 |tee -a spark-log.txt
```