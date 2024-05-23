# Pachyderm Snowflake Churn Risk Example

This example runs a Pachyderm pipeline on top of a Snowflake warehouse to compute the risk of
customer churn. The input is a sample from the Kaggle Telco Customer Churn dataset, you can get the
full dataset [here](https://www.kaggle.com/datasets/blastchar/telco-customer-churn).

# Setup

Before you start you'll need a Snowflake warehouse and a Pachyderm cluster. Create a database called
`churn_demo` and in it a table called `customer` to hold your customer data. This can be done by
running the provided `data.sql` file in the Snowflake console.

# Deploying the pipeline

## Credentials

In order for the pipeline to access Snowflake it'll need credentials. You can create one in
kubernetes with the following command:

```sh
kubectl create secret generic snowflakesecret --type=string \
--from-literal=PACHYDERM_SQL_PASSWORD='<password>'
```

## Ingest pipeline

The ingest pipeline ingest data from Snowflake into Pachyderm you can deploy it with:

```sh
pachctl update pipeline --jsonnet src/templates/sql_ingest_cron.jsonnet  \
--arg name=customer \
--arg
url="snowflake://<username>@<account>/<database>/PUBLIC?warehouse=COMPUTE_WH"
\
--arg query="select * from churn_demo.public.customer" \
--arg cronSpec="@every 30s" \
--arg secretName="snowflakesecret" \
--arg format=csv
```

Note that this pipeline will ingest from Snowflake every 30s, this is good for demos because it
gives immediate results, but can be a resource drain if it's left running overnight.

## Churn pipeline

The churn pipeline is implemented in Python. You can view the code and the Dockerfile that implement
it in `churn.py` and `Dockerfile` you don't need to do anything with those files because we've
already built and pushed the image as `pachyderm/churn`. The pipeline spec is `churn.pipeline.json`
to create the pipeline do:

```sh
pachctl create pipeline -f churn.pipeline.json
```

The pipeline spec should look pretty familiar to anyone who's created a Pachyderm pipeline before
with the exception of the egress section. The egress section defines a datawarehouse to egress the
pipeline's output to. Pachyderm lets you specify the format of your data, in this case it's CSV,
which means it expects directories in `/pfs/out` corresponding to tables you want to egress to with
`.csv` files inside them containing rows to egress.

# Conclusion

Once everything is deployed you have a production pipeline which will consume data from Snowflake on
a set schedule, process it and push it into another Snowflake table. You can join the input and
output table together to get an augmented table with risk of user churn added in. As new data lands
in the input table it will be picked up and reflected in the output table.
