"""
This is a reproduction of pachyderm's opencv example in python.
A walk-through is available in the pachyderm docs:
  https://docs.pachyderm.io/en/latest/getting_started/beginner_tutorial.html

It makes heavy use of python_pachyderm's higher-level utility functionality
 (`create_python_pipeline`, `put_files`), as well as more run-of-the-mill
 functionality (`create_repo`, `create_pipeline`).
"""
import os
import shutil
import tempfile
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps


def main(client: Client):
    # Create a repo called images
    images = pfs.Repo.from_uri("images")
    client.pfs.create_repo(repo=images)

    # Create the edges pipeline (and the edges repo automatically). This
    # pipeline runs when data is committed to the images repo, as indicated
    # by the input field.
    edges = pps.Pipeline(name="edges")
    client.pps.create_pipeline(
        pipeline=edges,
        transform=pps.Transform(
            cmd=["python3", "/edges.py"],
            image="pachyderm/opencv",
        ),
        input=pps.Input(pfs=pps.PfsInput(repo=images.name, glob="/*")),
    )

    # Create the montage pipeline (and the montage repo automatically). This
    # pipeline runs when data is committed to either the images repo or edges
    # repo, as indicated by the input field.
    client.pps.create_pipeline(
        pipeline=pps.Pipeline(name="montage"),
        transform=pps.Transform(
            cmd=["sh"],
            image="v4tech/imagemagick",
            stdin=[
                "montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs ! -name .env -type f | sort) /pfs/out/montage.png"
            ],
        ),
        input=pps.Input(
            cross=[
                pps.Input(pfs=pps.PfsInput(glob="/", repo=images.name)),
                pps.Input(pfs=pps.PfsInput(glob="/", repo=edges.name)),
            ]
        ),
    )

    with client.pfs.commit(branch=pfs.Branch.from_uri("images@master")) as commit:
        # Add some images from urls.
        # Alternatively, you could use `client.put_file_from_file` or
        # `client_put_file_bytes`.
        client.pfs.put_file_from_url(commit=commit, path="/liberty.jpg", url="https://docs.pachyderm.com/images/opencv/liberty.jpg")
        client.pfs.put_file_from_url(commit=commit, path="/kitten.jpg", url="https://docs.pachyderm.com/images/opencv/kitten.jpg")
        client.pfs.put_file_from_url(commit=commit, path="/robot.jpg", url="https://docs.pachyderm.com/images/opencv/robot.jpg")

    # Wait for the commit (and its downstream commits) to finish
    commit.wait_set()

    job = pps.Job(pipeline=pps.Pipeline(name="montage"), id=commit.id)
    if client.pps.inspect_job(job=job).state != pps.JobState.JOB_SUCCESS:
        print("Montage job failed, aborting. Check the pipeline logs for more details:")
        print("pachctl logs --pipeline=montage")
        exit(1)

    # Get the montage
    source_file = client.pfs.pfs_file(file=pfs.File.from_uri("montage@master:/montage.png"))
    with tempfile.NamedTemporaryFile(suffix="montage.png", delete=False) as dest_file:
        shutil.copyfileobj(source_file, dest_file)
        print("montage written to {}".format(dest_file.name))


def clean(client: Client):
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="montage"))
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="edges"))
    client.pfs.delete_repo(repo=pfs.Repo.from_uri("images"), force=True)


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

    clean(client)
    main(client)
