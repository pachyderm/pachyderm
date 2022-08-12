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

### Options from Clayton
# conf.set("spark.hadoop.fs.s3a.path.style.access", 'false') <-- DO NOT SET THIS TO FALSE, IT MUST BE SET TO TRUE (SEE ABOVE)
# conf.set("spark.hadoop.fs.s3a.connection.ssl.enabled", True)
# Clayton: "There are no other impls to try"
# conf.set("spark.hadoop.fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem")

# Setting this to 'none' should disable change detection (and thus etags)
# completely
# Clayton: "Same issues, with or without this config"
conf.set("spark.hadoop.fs.s3a.change.detection.mode", 'none')
# This might also disable change detection ("Determines if s3 object version
# attribute defined by fs.s3a.change.detection.source should be treated as
# 'required'")j
conf.set("spark.hadoop.fs.s3a.change.detection.version.required", 'false')
# conf.set("spark.hadoop.com.amazonaws.services.s3.enableV4", True)k


sc = SparkContext(conf=conf)
sc.setLogLevel("DEBUG")
sc.setSystemProperty("com.amazonaws.services.s3.disablePutObjectMD5Validation", "true")

# confirm config is applied to this session
spark = SparkSession.builder.getOrCreate()
conf = spark.sparkContext.getConf()
sc = spark.sparkContext
conf = sc.getConf()
print(sc.getConf().getAll())

# create some silly data
df = spark.createDataFrame([ Row(a=1, b=2.,) ])
df.show()

# df.write.parquet('s3a://master.rando2/nonemptyprefix6', mode="overwrite")
df.coalesce(1).write.format("parquet").mode("overwrite").save("s3a://master.rando2/nonemptyprefix8")

