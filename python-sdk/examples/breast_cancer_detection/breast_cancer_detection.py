"""
This is a reproduction of pachyderm's breast cancer detection example in 
python:
  https://github.com/pachyderm/examples/tree/master/breast-cancer-detection

This example is intended to be run from the 
pachyderm/python-sdk/examples/breast_cancer_detection directory.
"""
import os

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps


def main(client: Client):
    # Create repos for models and sample data
    models = pfs.Repo.from_uri("models")
    client.pfs.create_repo(repo=models)

    sample_data = pfs.Repo.from_uri("sample_data")
    client.pfs.create_repo(repo=sample_data)

    # Populate the repos with data and commit
    models_src_path = os.getcwd() + "/models"
    with client.pfs.commit(branch=pfs.Branch.from_uri("models@master")) as commit:
        client.pfs.put_files(commit=commit, source=models_src_path, path="/")
    commit.wait_all()

    sample_data_src_path = os.getcwd() + "/sample_data"
    with client.pfs.commit(branch=pfs.Branch.from_uri("sample_data@master")) as commit:
        client.pfs.put_files(commit=commit, source=sample_data_src_path, path="/")
    commit.wait_all()

    # Create image cropping pipeline
    crop = pps.Pipeline(name="crop")
    client.pps.create_pipeline(
        pipeline=crop,
        description="Remove background of image and save cropped files.",
        transform=pps.Transform(
            cmd=["/bin/bash", "multi-stage/crop.sh"],
            image="pachyderm/breast_cancer_classifier:1.11.6",
        ),
        input=pps.Input(pfs=pps.PfsInput(repo=sample_data.name, glob="/*")),
    )

    # Create pipeline to extract image centers frop cropped images
    extract_centers = pps.Pipeline(name="extract_centers")
    client.pps.create_pipeline(
        pipeline=extract_centers,
        description="Compute and extract optimal image centers.",
        transform=pps.Transform(
            cmd=["/bin/bash", "multi-stage/extract_centers.sh"],
            image="pachyderm/breast_cancer_classifier:1.11.6",
        ),
        input=pps.Input(pfs=pps.PfsInput(repo=crop.name, glob="/*")),
    )

    # Create heatmap generation pipeline that uses both pipeline outputs (as well as models)
    generate_heatmaps = pps.Pipeline(name="generate_heatmaps")
    client.pps.create_pipeline(
        pipeline=generate_heatmaps,
        description="Generates benign and malignant heatmaps for cropped images using patch classifier.",
        transform=pps.Transform(
            cmd=["/bin/bash", "multi-stage/generate_heatmaps.sh"],
            image="pachyderm/breast_cancer_classifier:1.11.6",            
        ),
        input=pps.Input(
            cross=[
                pps.Input(
                    join=[
                        pps.Input(pfs=pps.PfsInput(
                            repo=crop.name,
                            glob="/(*)",
                            join_on="$1",
                            lazy=False,
                        )),
                        pps.Input(pfs=pps.PfsInput(
                            repo=extract_centers.name,
                            glob="/(*)",
                            join_on="$1",
                            lazy=False,
                        )),
                    ],
                ),
                pps.Input(pfs=pps.PfsInput(
                        repo=models.name,
                        glob="/",
                        lazy=False,
                    ),
                ),
            ],
        ),
        resource_limits=pps.ResourceSpec(
            gpu=pps.GpuSpec(
                type="nvidia.com/gpu",
                number=1,
            ),
        ),
        resource_requests=pps.ResourceSpec(
            memory="4G",
            cpu=1.0,
        ),
    )

    # Create classification pipeline, using all 3 pipeline outputs and model
    classify = pps.Pipeline(name="classify")
    client.pps.create_pipeline(
        pipeline=classify,
        description="Runs the image only model and image+heatmaps model for breast cancer prediction.",
        transform=pps.Transform(
            cmd=["/bin/bash", "multi-stage/classify.sh"],
            image="pachyderm/breast_cancer_classifier:1.11.6", 
        ),
        input=pps.Input(
            cross=[
                pps.Input(
                    join=[
                        pps.Input(pfs=pps.PfsInput(
                            repo=crop.name,
                            glob="/(*)",
                            join_on="$1",
                        )),
                        pps.Input(pfs=pps.PfsInput(
                            repo=extract_centers.name,
                            glob="/(*)",
                            join_on="$1",
                        )),
                        pps.Input(pfs=pps.PfsInput(
                            repo=generate_heatmaps.name,
                            glob="/(*)",
                            join_on="$1",
                        )),
                    ],
                ),
                pps.Input(pfs=pps.PfsInput(
                        repo=models.name,
                        glob="/"
                    ),
                ),
            ],
        ),
        resource_limits=pps.ResourceSpec(
            gpu=pps.GpuSpec(
                type="nvidia.com/gpu",
                number=1,
            ),
        ),
        resource_requests=pps.ResourceSpec(
            memory="4G",
            cpu=1.0,
        ),
    )


def clean(client: Client):
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="classify"))
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="generate_heatmaps"))
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="extract_centers"))
    client.pps.delete_pipeline(pipeline=pps.Pipeline(name="crop"))
    client.pfs.delete_repo(repo=pfs.Repo.from_uri("sample_data"), force=True)
    client.pfs.delete_repo(repo=pfs.Repo.from_uri("models"), force=True)


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
