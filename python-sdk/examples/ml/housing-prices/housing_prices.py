# TODO: Recreates automated deferred processing example in python but needs
# polish before ready to publish as example
from dataclasses import fields
import json

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps


def main(client: Client):
    # Step 1: Create input data repository
    client.pfs.create_repo(repo=pfs.Repo(name="housing_data"))

    # Step 2: Create the regression pipeline
    spec = json.load(open("regression.json"))
    obj = pps.CreatePipelineRequest().from_dict(spec)
    client.pps.create_pipeline(**{field.name: getattr(obj, field.name) for field in fields(obj)})

    # Step 3: Add the housing dataset to the repo
    with client.pfs.commit(branch=pfs.Branch.from_uri("housing_data@master")) as c:
        with open("data/housing-simplified-1.csv", "rb") as file:
            c.put_file_from_file(path="housing-simplified.csv", file=file)

    # Step 4: Download files once the pipeline has finished
    client.pfs.wait_commit(pfs.Commit.from_uri("regression@master"))
    with client.pfs.pfs_tar_file(file=pfs.File.from_uri("regression@master:/")) as pfs_tar_file:
        pfs_tar_file.extractall(".")

    # Step 5: Update dataset with more data
    with client.pfs.commit(branch=pfs.Branch.from_uri("housing_data@master")) as c:
        with open("data/housing-simplified-2.csv", "rb") as file:
            c.put_file_from_file(path="housing-simplified.csv", file=file)

    # Step 6: Inspect the lineage of the pipeline
    print(list(client.pfs.list_commit(repo=pfs.Repo(name="regression"))))


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
