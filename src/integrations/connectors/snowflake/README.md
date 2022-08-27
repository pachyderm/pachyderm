# Snowflake Connector

## Read

```
pachctl update pipeline --jsonnet src/integrations/connectors/snowflake/snowflake-read.jsonnet \
    --arg image='pachyderm/snowflake:local' \
    --arg debug=true \
    --arg name=read \
    --arg cronSpec="@yearly" \
    --arg account='WDEUWUD-CJ80497' \
    --arg user='CIRCLECI' \
    --arg role='ROBOT' \
    --arg warehouse='COMPUTE_WH' \
    --arg database='TEST_82DCB7C6D25221F1' \
    --arg schema='public' \
    --arg query='select * from test_table' \
    --arg fileFormat="type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22' COMPRESSION=NONE" \
    --arg partitionBy='to_varchar(C_ID)'
```

TODO
- support user defined output file or output directory

## Write

```
pachctl update pipeline --jsonnet src/integrations/connectors/snowflake/snowflake-write.jsonnet \
    --arg image='pachyderm/snowflake:local' \
    --arg debug=true \
    --arg name=write \
    --arg inputRepo='read' \
    --arg account='WDEUWUD-CJ80497' \
    --arg user='CIRCLECI' \
    --arg role='ROBOT' \
    --arg warehouse='COMPUTE_WH' \
    --arg database='TEST_7CA7F2DD977D0BF0' \
    --arg schema='PUBLIC' \
    --arg table='test_table' \
    --arg fileFormat="type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22'"
```

TODO
- add support for VALIDATION_MODE in `COPY INTO <table>` query
- support upserts without dropping table every time
