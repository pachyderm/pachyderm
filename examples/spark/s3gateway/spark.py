from pyspark.sql import SparkSession, Row, DataFrame
from pyspark.context import SparkContext
from pyspark import SparkConf
import time
import os
import python_pachyderm

conf = SparkConf()
minio = True
if minio:
    conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:9000")
else:
    conf.set('spark.hadoop.fs.s3a.endpoint', "http://192.168.49.2:30600")
    # conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:30600")

conf.set('spark.hadoop.fs.s3a.impl', "org.apache.hadoop.fs.s3a.S3AFileSystem")

# XXX I don't think the following line actually turns on the magic committer. What else needs to happen?
# conf.set('spark.hadoop.fs.s3a.committer.name', 'magic')

if minio:
    conf.set('spark.hadoop.fs.s3a.access.key', 'admin')
    conf.set('spark.hadoop.fs.s3a.secret.key', 'password')
else:
    conf.set('spark.hadoop.fs.s3a.access.key', 'anything_matching')
    conf.set('spark.hadoop.fs.s3a.secret.key', 'anything_matching')

conf.set('spark.hadoop.fs.s3a.path.style.access', 'true')
conf.set('spark.hadoop.fs.s3a.connection.ssl.enabled', 'false')
conf.set("spark.hadoop.fs.s3a.change.detection.mode", 'none')
conf.set("spark.hadoop.fs.s3a.change.detection.version.required", 'false')

sc = SparkContext(conf=conf)
sc.setLogLevel("ERROR")
# sc.setLogLevel("DEBUG")
sc.setSystemProperty("com.amazonaws.services.s3.disablePutObjectMD5Validation", "true")

# confirm config is applied to this session
spark = SparkSession.builder.getOrCreate()
conf = spark.sparkContext.getConf()
sc = spark.sparkContext
conf = sc.getConf()
print(sc.getConf().getAll())

# create some example data
# big = "INFINITEIMPROBABILITY"*1024*100
big = "INFINITEIMPROBABILITY"*1024*100
zs = [ Row(a=big, b=big,) for _ in range(1000) ]
df = spark.createDataFrame(zs)
df.explain()
# df.show()
df.repartition(200)
df.explain()

repo = "spark-s3g-demo2"
branch = "master"

client = python_pachyderm.Client()

with client.commit(repo, branch) as commit:
    print(f"Opening commit {commit} for spark job")
    path = "example-data-21"
    if minio:
        url = f"s3a://foo/{path}"
    else:
        url = f"s3a://{branch}.{repo}/{path}"
    df.coalesce(1).write.format("parquet").mode("overwrite").save(url)
    df.explain()
    print(f"Closing {commit}")