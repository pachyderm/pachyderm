from pyspark.sql import SparkSession


spark = SparkSession.builder \
            .master("local[1]") \
            .appName("my_app") \
            .getOrCreate()

spark._jsc.hadoopConfiguration().set("fs.s3a.endpoint", "http://localhost:30600")
# spark._jsc.hadoopConfiguration().set("fs.s3a.access.key", "foo")
# spark._jsc.hadoopConfiguration().set("fs.s3a.secret.key", "foo")
spark._jsc.hadoopConfiguration().set("fs.s3a.impl","org.apache.hadoop.fs.s3a.S3AFileSystem")
spark._jsc.hadoopConfiguration().set("fs.s3a.connection.ssl.enabled", "false")
spark._jsc.hadoopConfiguration().set("fs.s3a.metadatastore.impl", "org.apache.hadoop.fs.s3a.s3guard.NullMetadataStore")
spark._jsc.hadoopConfiguration().set("fs.s3a.path.style.access", "true")

spark.sparkContext.setLogLevel("debug")


df = spark.read.format("json").load('s3a://subset-10.fhir/sample.json')
df.show()
