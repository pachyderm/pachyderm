from pyspark.sql import SparkSession, Row, DataFrame
from pyspark.context import SparkContext
from pyspark import SparkConf
import time
import os
# import python_pachyderm

conf = SparkConf()
minio = False
if minio:
    conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:9000")
else:
    # conf.set('spark.hadoop.fs.s3a.endpoint', "http://192.168.49.2:30600")
    endpoint = os.getenv('S3_ENDPOINT')
    conf.set('spark.hadoop.fs.s3a.endpoint', endpoint)
    print(f"endpoint is {endpoint}")
    # conf.set('spark.hadoop.fs.s3a.endpoint', "http://localhost:30600")

conf.set('spark.hadoop.fs.s3a.impl', "org.apache.hadoop.fs.s3a.S3AFileSystem")

# XXX I don't think the following line actually turns on the magic committer. What else needs to happen?
# conf.set('spark.hadoop.fs.s3a.committer.name', 'magic')

if minio:
    conf.set('spark.hadoop.fs.s3a.access.key', 'admin')
    conf.set('spark.hadoop.fs.s3a.secret.key', 'password')
else:
    # TODO: make it so we don't need these
    conf.set('spark.hadoop.fs.s3a.access.key', os.getenv("AWS_ACCESS_KEY_ID"))
    conf.set('spark.hadoop.fs.s3a.secret.key', os.getenv("AWS_SECRET_ACCESS_KEY"))

conf.set('spark.hadoop.fs.s3a.path.style.access', 'true')
conf.set('spark.hadoop.fs.s3a.connection.ssl.enabled', 'false')
conf.set("spark.hadoop.fs.s3a.change.detection.mode", 'none')
conf.set("spark.hadoop.fs.s3a.change.detection.version.required", 'false')

sc = SparkContext(conf=conf)
# sc.setLogLevel("ERROR")
sc.setLogLevel("DEBUG")
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

# client = python_pachyderm.Client()

# with client.commit(repo, branch) as commit:

# print(f"Opening commit {commit} for spark job")

path = "example-data-24"
if minio:
    url = f"s3a://foo/{path}"
else:
    # url = f"s3a://{branch}.{repo}/{path}"
    url = f"s3a://pachyderm-test/{path}"


print("Starting write...")
(df.coalesce(1)
    .write
#    .option("fs.s3a.committer.name", "magic")
    .format("parquet")
    .mode("overwrite")
    .save(url))
print("Finished write!")

df.explain()

# print(f"Closing {commit}")