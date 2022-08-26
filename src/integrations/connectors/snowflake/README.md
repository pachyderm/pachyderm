# Snowflake Connector

## Read

```
pachctl create pipeline --jsonnet src/integrations/connectors/snowflake/snowflake-read.jsonnet \
    --arg image='pachyderm/snowflake:local' \
    --arg name=read \
    --arg cronSpec="@yearly" \
    --arg account='WDEUWUD-CJ80497' \
    --arg user='CIRCLECI' \
    --arg role='ROBOT' \
    --arg warehouse='COMPUTE_WH' \
    --arg database='TEST_82DCB7C6D25221F1' \
    --arg schema='public' \
    --arg query='select * from test_table' \
    --arg fileFormat="(type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22' COMPRESSION=NONE)" \
    --arg partitionBy='to_varchar(C_ID)'
```

## Write

```
pachctl create pipeline --jsonnet src/integrations/connectors/snowflake/snowflake-write.jsonnet \
    --arg image=pachyderm/snowflake:local \
    --arg name=write \
    --arg inputRepo='read' \
    --arg account='WDEUWUD-CJ80497' \
    --arg user='CIRCLECI' \
    --arg role='ROBOT' \
    --arg warehouse='COMPUTE_WH' \
    --arg database='TEST_82DCB7C6D25221F1' \
    --arg schema='PUBLIC' \
    --arg fileFormat="(type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22')"
```