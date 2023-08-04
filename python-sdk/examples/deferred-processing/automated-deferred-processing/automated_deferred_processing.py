# TODO: Recreates automated deferred processing example in python but needs
# polish before ready to publish as example
from dataclasses import fields
import json

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps


def create_dag(client: Client):
    client.pfs.create_repo(repo=pfs.Repo(name="images_dp_1"))
    client.pfs.create_repo(repo=pfs.Repo(name="images_dp_2"))

    spec = json.load(open("../deferred_processing_plus_transactions/edges_dp.json"))
    obj = pps.CreatePipelineRequest().from_dict(spec)
    client.pps.create_pipeline(**{field.name: getattr(obj, field.name) for field in fields(obj)})
    
    spec = json.load(open("../deferred_processing_plus_transactions/montage_dp.json"))
    obj = pps.CreatePipelineRequest().from_dict(spec)
    client.pps.create_pipeline(**{field.name: getattr(obj, field.name) for field in fields(obj)})

    with client.pfs.commit(branch=pfs.Branch.from_uri("images_dp_1@master")) as c:
        with open("../deferred_processing_plus_transactions/images.txt", "rb") as file:
            c.put_file_from_file(path="/images.txt", file=file)
    with client.pfs.commit(branch=pfs.Branch.from_uri("images_dp_1@master")) as c:
        with open("../deferred_processing_plus_transactions/images2.txt", "rb") as file:
            c.put_file_from_file(path="/images2.txt", file=file)

    with client.pfs.commit(branch=pfs.Branch.from_uri("images_dp_2@master")) as c:
        with open("../deferred_processing_plus_transactions/images3.txt", "rb") as file:
            c.put_file_from_file(path="/images3.txt", file=file)


def if_no_auth(client: Client):
    # Step 1: Create deferred processing DAG
    create_dag(client)

    # Step 2: Create branch mover cron pipeline
    spec = json.load(open("branch-mover-no-auth.json"))
    obj = pps.CreatePipelineRequest().from_dict(spec)
    client.pps.create_pipeline(**{field.name: getattr(obj, field.name) for field in fields(obj)})

    # Step 3: Watch pachctl jobs using command `watch -cn 2 pachctl list job --no-pager`

    # Step 4: Commit data to see deferred processing
    with client.pfs.commit(branch=pfs.Branch.from_uri("images_dp_1@master")) as c:
        c.put_file_from_url(path="/1VqcWw9.jpg", url="http://imgur.com/1VqcWw9.jpg")


def if_yes_auth(client: Client):
    # Step 1: Create deferred processing DAG
    create_dag(client)

    # Step 2: Create K8 secret
    token = client.auth.get_robot_token(robot="deferred-example", ttl=2246400).token # 624h in seconds
    client.pps.create_secret(name="pachyderm-user-secret", data={"auth_token": token})

    # Step 3: Create branch mover cron pipeline
    spec = json.load(open("branch-mover.json"))
    obj = pps.CreatePipelineRequest().from_dict(spec)
    client.pps.create_pipeline(**{field.name: getattr(obj, field.name) for field in fields(obj)})

    # Step 4: Watch pachctl jobs using command `watch -cn 2 pachctl list job --no-pager`

    # Step 5: Commit data to see deferred processing
    with client.pfs.commit(branch=pfs.Branch.from_uri("images_dp_1@master")) as c:
        c.put_file_from_url(path="/1VqcWw9.jpg", url="http://imgur.com/1VqcWw9.jpg")


def main(client: Client):
    if_no_auth(client)
    # if_yes_auth(client)


if __name__ == "__main__":
    # Connects to a pachyderm cluster using the pachctl config file located
    # at ~/.pachyderm/config.json. For other setups, you'll want one of the 
    # alternatives:
    # 1) To connect to pachyderm when this script is running inside the
    #    cluster, use `Client.new_in_cluster()`.
    # 2) To connect to pachyderm via a pachd address, use
    #    `Client.new_from_pachd_address`.
    # 3) To explicitly set the host and port, pass parameters into
    #    `Client()`.
    # 4) To use a config file located elsewhere, pass in the path to that
    #    config file to Client.from_config()
    client = Client.from_config()
    main(client)
