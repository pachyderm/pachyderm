# Periodic ingress from MongoDB

This example pipeline executes a query periodically against a MongoDB database outside of Pachyderm.  The results of the query are stored in a corresponding output repository.  This repository could be used to drive additional pipeline stages periodically based on the results of the query.

The example assumes that you have:

- A Pachyderm cluster running - see [Local Installation](https://docs.pachyderm.com/latest/getting_started/local_installation/) to get up and running with a local Pachyderm cluster in just a few minutes.
- The `pachctl` CLI tool installed and connected to your Pachyderm cluster - see [deploy docs](https://docs.pachyderm.com/latest/deploy-manage/) for instructions.

## Setup MongoDB

The easiest way to demonstrate this example is with a free hosted MongoDB cluster, such as the free tier of [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) or [MLab](https://mlab.com/) (although you could certainly do this with any MongoDB). Assuming that you are using MongoDB Atlas:

1. Deploy a new `Cluster0` MongoDB cluster with MongoDB Atlas (remmeber the admin username and password as you will need these shortly).  Once deployed you should be able to see this cluster in the MongoDB Atlas dashboard:

![alt text](mongo1.png)

1. Click on the "connect" button for your cluster and make sure that all IPs are whitelisted (or at least the k8s master IP where you have Pachyderm deployed):

![alt text](mongo2.png)

1. Then click on "Connect with the MongoDB shell" to find the URI, DB name (`test` if you are using MongoDB Atlas `Cluster0`), username, and authentication DB for connecting to your cluster.  You will need these to query MongoDB.

1. Make sure you have the MongoDB tools installed locally. You can follow [this guide](https://docs.mongodb.com/manual/administration/install-community/) to install themk.

## Import example data

We are going to run this example with an example set of data about restaurants.  This dataset comes directly from MongoDB and is used in many of their examples as well.

1. Download the dataset from [here](https://raw.githubusercontent.com/mongodb/docs-assets/primer-dataset/primer-dataset.json).  It is named `primer-dataset.json`.  Each of the records in this dataset look like the following:

    ```
    {
      "address": {
         "building": "1007",
         "coord": [ -73.856077, 40.848447 ],
         "street": "Morris Park Ave",
         "zipcode": "10462"
      },
      "borough": "Bronx",
      "cuisine": "Bakery",
      "grades": [
         { "date": { "$date": 1393804800000 }, "grade": "A", "score": 2 },
         { "date": { "$date": 1378857600000 }, "grade": "A", "score": 6 },
         { "date": { "$date": 1358985600000 }, "grade": "A", "score": 10 },
         { "date": { "$date": 1322006400000 }, "grade": "A", "score": 9 },
         { "date": { "$date": 1299715200000 }, "grade": "B", "score": 14 }
      ],
      "name": "Morris Park Bake Shop",
      "restaurant_id": "30075445"
    }
    ```

1. Import the dataset to the `restaurants` collection in MongoDB (in the `test` DB if you are using MongoDB Atlas) using the `mongoimport` command.  You will need to specify the Mongo hosts, username, password, etc. from your MongoDB cluster.  For example:

    ```sh
    $ mongoimport --host Cluster0-shard-0/cluster0-shard-00-00-cwehf.mongodb.net:27017,cluster0-shard-00-01-cwehf.mongodb.net:27017,cluster0-shard-00-02-cwehf.mongodb.net:27017 --ssl -u admin -p '<my password>' --authenticationDatabase admin --db test --collection restaurants --drop --file primer-dataset.json         
    2017-08-28T13:40:38.983-0400    connected to: Cluster0-shard-0/cluster0-shard-00-00-cwehf.mongodb.net:27017,cluster0-shard-00-01-cwehf.mongodb.net:27017,cluster0-shard-00-02-cwehf.mongodb.net:27017
    2017-08-28T13:40:39.048-0400    dropping: test.restaurants
    2017-08-28T13:40:41.310-0400    [#.......................] test.restaurants     540KB/11.3MB (4.7%)
    2017-08-28T13:40:44.310-0400    [##......................] test.restaurants     1.04MB/11.3MB (9.2%)
    2017-08-28T13:40:47.310-0400    [##......................] test.restaurants     1.04MB/11.3MB (9.2%)
    2017-08-28T13:40:50.310-0400    [###.....................] test.restaurants     1.55MB/11.3MB (13.7%)
    2017-08-28T13:40:53.310-0400    [####....................] test.restaurants     2.07MB/11.3MB (18.2%)
    2017-08-28T13:40:56.310-0400    [#####...................] test.restaurants     2.58MB/11.3MB (22.8%)
    2017-08-28T13:40:59.310-0400    [######..................] test.restaurants     3.10MB/11.3MB (27.4%)
    2017-08-28T13:41:02.310-0400    [######..................] test.restaurants     3.10MB/11.3MB (27.4%)
    2017-08-28T13:41:05.310-0400    [#######.................] test.restaurants     3.61MB/11.3MB (31.9%)
    2017-08-28T13:41:08.310-0400    [########................] test.restaurants     4.12MB/11.3MB (36.4%)
    2017-08-28T13:41:11.310-0400    [#########...............] test.restaurants     4.64MB/11.3MB (41.0%)
    2017-08-28T13:41:14.310-0400    [#########...............] test.restaurants     4.64MB/11.3MB (41.0%)
    2017-08-28T13:41:17.310-0400    [##########..............] test.restaurants     5.15MB/11.3MB (45.5%)
    2017-08-28T13:41:20.310-0400    [############............] test.restaurants     5.67MB/11.3MB (50.0%)
    2017-08-28T13:41:23.310-0400    [#############...........] test.restaurants     6.19MB/11.3MB (54.6%)
    2017-08-28T13:41:26.310-0400    [#############...........] test.restaurants     6.19MB/11.3MB (54.6%)
    2017-08-28T13:41:29.310-0400    [##############..........] test.restaurants     6.70MB/11.3MB (59.2%)
    2017-08-28T13:41:32.310-0400    [###############.........] test.restaurants     7.21MB/11.3MB (63.7%)
    2017-08-28T13:41:35.310-0400    [################........] test.restaurants     7.71MB/11.3MB (68.1%)
    2017-08-28T13:41:38.310-0400    [################........] test.restaurants     7.71MB/11.3MB (68.1%)
    2017-08-28T13:41:41.310-0400    [#################.......] test.restaurants     8.18MB/11.3MB (72.3%)
    2017-08-28T13:41:44.310-0400    [##################......] test.restaurants     8.62MB/11.3MB (76.1%)
    2017-08-28T13:41:47.310-0400    [###################.....] test.restaurants     9.03MB/11.3MB (79.7%)
    2017-08-28T13:41:50.310-0400    [###################.....] test.restaurants     9.41MB/11.3MB (83.1%)
    2017-08-28T13:41:53.310-0400    [####################....] test.restaurants     9.77MB/11.3MB (86.3%)
    2017-08-28T13:41:56.310-0400    [#####################...] test.restaurants     10.1MB/11.3MB (89.2%)
    2017-08-28T13:41:59.310-0400    [######################..] test.restaurants     10.4MB/11.3MB (91.9%)
    2017-08-28T13:42:02.310-0400    [######################..] test.restaurants     10.7MB/11.3MB (94.5%)
    2017-08-28T13:42:05.310-0400    [#######################.] test.restaurants     11.0MB/11.3MB (97.0%)
    2017-08-28T13:42:08.310-0400    [########################] test.restaurants     11.3MB/11.3MB (100.0%)
    2017-08-28T13:42:08.449-0400    [########################] test.restaurants     11.3MB/11.3MB (100.0%)
    2017-08-28T13:42:08.449-0400    imported 25359 documents
    ```

## Create a Kubernetes secret with your Mongo creds 

In order for your Pachyderm pipeline to talk with MongoDB, we need to tell Pachyderm about the MongoDB URI, username, password, etc.  We will do this via a [Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/),
loaded via `pachctl create secret`.


1. The next few steps show you how to add a secret with the following five keys

   * `uri`
   * `username`
   * `password`
   * `db`
   * `collection`

   First, we'll save some values to files. 
   The values should all be enclosed in single quotes to prevent the shell from interpreting them.
   
   ```sh
   $ echo -n '<uri>' > uri ; chmod 600 uri
   $ echo -n '<username>' > username ; chmod 600 username
   $ echo -n '<password>' > password ; chmod 600 password
   $ echo -n '<db>' > db ; chmod 600 db
   $ echo -n '<collection>' > collection ; chmod 600 collection
   ```
   
1. Confirm the values in these files are what you expect.

   ```sh
   $ cat uri
   $ cat username
   $ cat password
   $ cat db
   $ cat collection
   ```
   
   Creating the secret will require different steps,
   depending on whether you have Kubernetes access or not.
   Pachyderm Hub users don't have access to Kubernetes.
   If you have Kubernetes access and want to use `kubectl`, 
   you may follow the two steps prefixed with "(Kubernetes)".
   If you don't have access to Kubernetes or don't want to use `kubectl`,
   follow the two steps labeled "(Pachyderm Hub)" 

1. (Kubernetes) If you have direct access to the Kubernetes cluster, you can create a secret using `kubectl`.
   
   ```sh
   $ kubectl create secret generic mongosecret --from-file=./uri \
       --from-file=./username \
       --from-file=./password \
       --from-file=./db \
       --from-file=./collection
   ```
   
1. (Kubernetes) Confirm that the secrets got set correctly.
   You use `kubectl get secret` to output the secrets, and then decode them using `jq` to confirm they're correct.
   
   ```sh
   $ kubectl get secret mongosecret -o json | jq '.data | map_values(@base64d)'
   {
       "uri": "<uri>",
       "username": "<username>"
       "password": "<password>"
       "db": "<db>"
       "collection": "<collection>"
   }
   ```

   You will have to use pachctl if you're using Pachyderm Hub,
   or don't have access to the Kubernetes cluster.
   The next three steps show how to do that.

1. (Pachyderm Hub) Create a secrets file from the provided template.
   
   ```sh
   $ jq -n --arg uri $(cat uri) --arg username $(cat username) \
       --arg password $(cat password) --arg db $(cat db) --arg collection $(cat collection) \
       -f mongodb-credentials-template.jq  > mongodb-credentials-secret.json 
   $ chmod 600 mongodb-credentials-secret.json
   ```

1. (Pachyderm Hub) Confirm the secrets file is correct by decoding the values.
   
   ```sh
   $ jq '.data | map_values(@base64d)' mongodb-credentials-secret.json
   {
       "uri": "<uri>",
       "username": "<username>"
       "password": "<password>"
       "db": "<db>"
       "collection": "<collection>"
   }
   ```

1. (Pachyderm Hub) Generate a secret using pachctl

   ```sh
   $ pachctl create secret -f mongodb-credentials-secret.json
   ```

## Create the pipeline, view the results

In our [pipeline spec](query.json), we will do the following:

- Grab the Kubernetes secret that we defined above.
- Define a `cron` input that will cause the pipeline to be triggered every 10 seconds.
- Using the official `mongo` Docker image, query for a random document from the `restaurants` collection and output that to `/pfs/out`.

```
{
  "pipeline": {
    "name": "query"
  },
  "transform": {
    "image": "mongo",
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "export uri=$(cat /tmp/mongosecret/uri)",
      "export db=$(cat /tmp/mongosecret/db)",
      "export collection=$(cat /tmp/mongosecret/collection)",
      "export username=$(cat /tmp/mongosecret/username)",
      "export password=$(cat /tmp/mongosecret/password)",
      "mongo \"$uri\" --authenticationDatabase admin --ssl --username $username --password $password --quiet --eval 'db.restaurants.aggregate({ $sample: { size: 1 } });' | tail -n1 | egrep -v \"^>|^bye\" > /pfs/out/output.json"
    ],
    "secrets": [ 
      {
        "name": "mongosecret",
        "mount_path": "/tmp/mongosecret"
      } 
    ]
  },
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 10s"
    }  
  }
}
```

This will allow us to view the head of the output over time to see a bunch of random documents being queried out of MongoDB.

1. Create the pipeline.

    ```sh
    $ pachctl create pipeline -f query.json
    ``` 
    
1. Run the command `pachctl list pipeline` to make sure it's running:
   
   ```sh
   $ pachctl list pipeline
   NAME      VERSION INPUT                     CREATED       STATE / LAST JOB   DESCRIPTION                                                                                               
   query     1       tick:@every 10s           6 seconds ago running / starting                                                                                    
   ```
   
1. After the pipeline is running, you should see jobs start to be triggered every 10 seconds.

    ```sh
    $ pachctl list job
    ID                                   OUTPUT COMMIT STARTED      DURATION RESTART PROGRESS  DL UL STATE            
    842e4e6c-4920-42c0-9c81-e5299b67e4a0 query/-       1 second ago -        0       0 + 0 / 1 0B 0B running 
    $ pachctl list job
    ID                                   OUTPUT COMMIT                          STARTED        DURATION  RESTART PROGRESS  DL  UL   STATE            
    5938a0d0-9512-455f-a390-14adc3669e5f query/0f8a2ba1150a463299ee71961427bdcb 3 seconds ago  3 seconds 0       1 + 0 / 1 26B 617B success 
    952427a6-c92d-4c98-a781-87616988d528 query/33776e4df3b24ab68d70b5185eb37661 13 seconds ago 1 second  0       1 + 0 / 1 26B 613B success 
    1bc5f608-85fd-44eb-833e-562d15629706 query/6dd2a4da566f4d30ad9c66fc60244bab 23 seconds ago 1 second  0       1 + 0 / 1 26B 721B success 
    efa677a4-7f83-424b-879d-70a0c5690bb2 query/f56b1f314030455c8bdf8a10b68ebd16 33 seconds ago 1 second  0       1 + 0 / 1 26B 529B success 
    842e4e6c-4920-42c0-9c81-e5299b67e4a0 query/2a11bfc3e6d74af0a8d254d3ecf6f6af 43 seconds ago 1 second  0       1 + 0 / 1 26B 535B success 
    $ pachctl list job
    ID                                   OUTPUT COMMIT                          STARTED            DURATION  RESTART PROGRESS  DL  UL   STATE            
    7ab2b1cf-bd13-4aa7-bf7b-b06f2c29242a query/-                                1 second ago       -         0       0 + 0 / 1 0B  0B   running 
    bc71d40b-5b1c-474d-a24d-f7487e037cee query/997587e3bd794e2ea48890e6022434f4 11 seconds ago     2 seconds 0       1 + 0 / 1 26B 836B success 
    41d499ad-7ba2-4ba6-82b0-68a8d5e3e67a query/b8eae461132a49819e59b62c39e6b6eb 21 seconds ago     1 second  0       1 + 0 / 1 26B 669B success 
    5938a0d0-9512-455f-a390-14adc3669e5f query/0f8a2ba1150a463299ee71961427bdcb 31 seconds ago     3 seconds 0       1 + 0 / 1 26B 617B success 
    952427a6-c92d-4c98-a781-87616988d528 query/33776e4df3b24ab68d70b5185eb37661 41 seconds ago     1 second  0       1 + 0 / 1 26B 613B success 
    1bc5f608-85fd-44eb-833e-562d15629706 query/6dd2a4da566f4d30ad9c66fc60244bab 51 seconds ago     1 second  0       1 + 0 / 1 26B 721B success 
    efa677a4-7f83-424b-879d-70a0c5690bb2 query/f56b1f314030455c8bdf8a10b68ebd16 About a minute ago 1 second  0       1 + 0 / 1 26B 529B success 
    842e4e6c-4920-42c0-9c81-e5299b67e4a0 query/2a11bfc3e6d74af0a8d254d3ecf6f6af About a minute ago 1 second  0       1 + 0 / 1 26B 535B success
    ```

1. You can then observe the output changing over time.  You can watch it change with each query by executing:

    ```sh
    $ watch pachctl get file query@master:output.json
    ```
    
    Or you can look at individual results over time via the commit IDs:

    ```sh
    $ pachctl get file query@master:output.json
    { "_id" : ObjectId("59a455af69a077c0dc028410"), "address" : { "building" : "119", "coord" : [ -73.9784962, 40.6788476 ], "street" : "5 Avenue", "zipcode" : "11217" }, "borough" : "Brooklyn", "cuisine" : "Mexican", "grades" : [ { "date" : ISODate("2014-07-29T00:00:00Z"), "grade" : "B", "score" : 27 }, { "date" : ISODate("2014-03-10T00:00:00Z"), "grade" : "B", "score" : 15 }, { "date" : ISODate("2014-02-12T00:00:00Z"), "grade" : "P", "score" : 3 }, { "date" : ISODate("2013-09-05T00:00:00Z"), "grade" : "C", "score" : 35 }, { "date" : ISODate("2013-03-06T00:00:00Z"), "grade" : "A", "score" : 12 }, { "date" : ISODate("2012-09-12T00:00:00Z"), "grade" : "A", "score" : 13 }, { "date" : ISODate("2012-04-17T00:00:00Z"), "grade" : "A", "score" : 12 } ], "name" : "El Pollito Mexicano", "restaurant_id" : "41051406" }
    $ pachctl get file query@64ac2bd721d04212a3a0b90833f751e5:output.json
    { "_id" : ObjectId("59a455f069a077c0dc02e16e"), "address" : { "building" : "1650", "coord" : [ -73.928079, 40.856481 ], "street" : "Saint Nicholas Ave", "zipcode" : "10040" }, "borough" : "Manhattan", "cuisine" : "Spanish", "grades" : [ { "date" : ISODate("2015-01-20T00:00:00Z"), "grade" : "Not Yet Graded", "score" : 2 } ], "name" : "Angebienvendia", "restaurant_id" : "50018661" }
    $ pachctl get file query@74a6cf68de2047fe94ac7982065df03d:output.json
    { "_id" : ObjectId("59a455b669a077c0dc02904d"), "address" : { "building" : "14", "coord" : [ -73.990382, 40.741571 ], "street" : "West   23 Street", "zipcode" : "10010" }, "borough" : "Manhattan", "cuisine" : "Café/Coffee/Tea", "grades" : [ { "date" : ISODate("2014-05-02T00:00:00Z"), "grade" : "A", "score" : 11 }, { "date" : ISODate("2013-11-22T00:00:00Z"), "grade" : "A", "score" : 8 }, { "date" : ISODate("2012-11-20T00:00:00Z"), "grade" : "A", "score" : 9 }, { "date" : ISODate("2011-11-18T00:00:00Z"), "grade" : "A", "score" : 6 } ], "name" : "Starbucks Coffee (Store #13539)", "restaurant_id" : "41290548" }
    ```
