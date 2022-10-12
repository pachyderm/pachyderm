FROM jupyter/pyspark-notebook:spark-3.3.0
WORKDIR /home/jovyan
ADD spark.py /home/jovyan/
RUN wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.3/hadoop-aws-3.3.3.jar
RUN wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.264/aws-java-sdk-bundle-1.12.264.jar
ENTRYPOINT ["spark-submit", "--executor-memory=15g", "--driver-memory=15g", "--jars='hadoop-aws-3.3.3.jar,aws-java-sdk-bundle-1.12.264.jar'", "/home/jovyan/spark.py"]