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


def relpath(path):
    return os.path.join(os.path.dirname(os.path.abspath(__file__)), path)


def main():
    # Connects to a pachyderm cluster on the default host:port
    # (`localhost:30650`). This will work for certain environments (e.g. k8s
    # running on docker for mac), as well as when port forwarding is being
    # used. For other setups, you'll want one of the alternatives:
    # 1) To connect to pachyderm when this script is running inside the
    #    cluster, use `Client.new_in_cluster()`.
    # 2) To connect to pachyderm via a pachd address, use
    #    `Client.new_from_pachd_address`.
    # 3) To explicitly set the host and port, pass parameters into
    #   `Client()`.
    client = Client()

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
                "montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png"
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
        # Add some images, recursively inserting content from the images
        # directory. Alternatively, you could use `client.put_file_url` or
        # `client_put_file_bytes`.
        source = None  # TODO: Add correct path
        client.pfs.put_files(commit=commit, source=source, path="/")

    # Wait for the commit (and its downstream commits) to finish
    commit.wait_set()

    # Get the montage
    source_file = client.pfs.pfs_file(file=pfs.File.from_uri("montage@master:/montage.png"))
    with tempfile.NamedTemporaryFile(suffix="montage.png", delete=False) as dest_file:
        shutil.copyfileobj(source_file, dest_file)
        print("montage written to {}".format(dest_file.name))


def clean():
    client = Client()
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="montage"))
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="edges"))

    client.pfs.delete_repo(repo=pfs.Repo.from_uri("images"), force=True)


if __name__ == "__main__":
    # clean()
    main()
