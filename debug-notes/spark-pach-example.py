import os
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession

# os.environ["AWS_SECRET_ACCESS_KEY"] = "foo"
# os.environ["AWS_ACCESS_KEY_ID"] = "foo"

config = {
    # "spark.hadoop.fs.s3a.endpoint": "http://pachd.pachyderm-2-0.svc.cluster.local:30600",
    "spark.hadoop.fs.s3a.endpoint": "http://localhost:30600",
    "spark.hadoop.fs.s3a.connection.ssl.enabled": "false",
    "spark.hadoop.fs.s3a.path.style.access": "true",
    "spark.hadoop.fs.s3a.impl": "org.apache.hadoop.fs.s3a.S3AFileSystem",
}

def get_spark_session(app_name: str, conf: SparkConf):
    # conf.setMaster("k8s://https://kubernetes.default.svc.cluster.local")
    # conf.setMaster("k8s://https://192.168.39.149:8443")
    conf.setMaster("local")
    for key, value in config.items():
        conf.set(key, value)
    return SparkSession.builder.appName(app_name).config(conf=conf).getOrCreate()

spark = get_spark_session("my-spark-job", SparkConf())

df = spark.read.format("binaryFile").load('s3a://subset-10.fhir/')
print(df.show())
