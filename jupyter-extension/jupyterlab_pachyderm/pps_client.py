import json
import yaml
import os.path
from dataclasses import dataclass, asdict
from datetime import datetime
from inspect import getsource
from pathlib import Path
from textwrap import dedent
from typing import List, Optional, Union

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps
from nbconvert import PythonExporter
from tornado.web import HTTPError

from .log import get_logger

METADATA_KEY = "pachyderm_pps"


@dataclass
class PpsConfig:

    notebook_path: Path
    pipeline: pps.Pipeline
    image: str
    requirements: Optional[str]
    external_files: List[Path]
    port: str
    gpu_mode: str
    resource_spec: Optional[pps.ResourceSpec]
    input_spec: pps.Input

    @classmethod
    def from_notebook(cls, notebook_path: Union[str, Path]) -> "PpsConfig":
        """Parses a config from the metadata of a notebook file.

        Raises ValueError if required field is not specified.
        """
        notebook_path = Path(notebook_path)
        notebook_data = json.loads(notebook_path.read_bytes())

        metadata = notebook_data["metadata"].get(METADATA_KEY)
        if metadata is None:
            raise ValueError(f"{METADATA_KEY} not specified in notebook metadata")

        config = metadata.get("config")
        if config is None:
            raise ValueError(
                f"{METADATA_KEY}.config not specified in notebook metadata"
            )

        pipeline_data = config.get("pipeline")
        if pipeline_data is None:
            raise ValueError("field pipeline not set")

        if "project" in pipeline_data:
            project = pfs.Project(
                name=pipeline_data["project"].get("name") or "default"
            )
        else:
            project = pfs.Project(name="default")

        pipeline = pps.Pipeline(name=pipeline_data.get("name"), project=project)

        image = config.get("image")
        if image is None:
            raise ValueError("field image not set")

        requirements = config.get("requirements")
        if requirements:
            requirements = notebook_path.parent.joinpath(requirements).resolve()

        external_files = []
        external_files_str = config.get("external_files", "").strip()
        if external_files_str:
            for external_file in external_files_str.strip().split(','):
                external_files.append(notebook_path.with_name(Path(external_file.strip()).name))

        input_spec_str = config.get("input_spec")
        if input_spec_str is None:
            raise ValueError("field input_spec not set")
        input_spec_dict = yaml.safe_load(input_spec_str)
        if input_spec_dict is None:
            raise ValueError("invalid input spec")
        input_spec = pps.Input().from_dict(input_spec_dict)

        port = config.get("port")

        gpu_mode = config.get("gpu_mode")
        resource_spec_str = config.get("resource_spec")
        if not resource_spec_str:
            resource_spec = None
        else:
            resource_spec_dict = yaml.safe_load(resource_spec_str)
            resource_spec = pps.ResourceSpec().from_dict(resource_spec_dict)
        return cls(
            notebook_path=notebook_path,
            pipeline=pipeline,
            image=image,
            requirements=requirements,
            external_files=external_files,
            input_spec=input_spec,
            port=port,
            gpu_mode=gpu_mode,
            resource_spec=resource_spec,
        )

    def to_dict(self):
        data = asdict(self)
        del data["notebook_path"]
        if data['requirements'] is not None:
            data['requirements'] = str(data['requirements'])
        return data


def companion_repo_name(config: PpsConfig):
    return f"{config.pipeline.name}__context"


def create_pipeline_spec(config: PpsConfig, companion_branch: str) -> str:
    companion_repo = companion_repo_name(config)
    input_spec = pps.Input(
        cross=[
            config.input_spec,
            pps.Input(
                pfs=pps.PfsInput(
                    project=config.pipeline.project.name,
                    repo=companion_repo,
                    branch=companion_branch,
                    glob="/",
                )
            ),
        ]
    )
    pipeline_spec = pps.CreatePipelineRequest(
        pipeline=config.pipeline,
        description="Auto-generated from notebook",
        transform=pps.Transform(
            cmd=["python3", "-u", f"/pfs/{companion_repo}/entrypoint.py"],
            image=config.image,
        ),
        input=input_spec,
        update=True,
        reprocess=True,
    )
    if config.port:
        pipeline_spec.service = pps.Service(
            external_port=config.port, internal_port=config.port, type="LoadBalancer"
        )
    if config.gpu_mode == "Simple":
        pipeline_spec.resource_limits = pps.ResourceSpec(
            gpu=pps.GpuSpec(type="nvidia.com/gpu", number=1)
        )
    elif config.gpu_mode == "Advanced":
        pipeline_spec.resource_limits = config.resource_spec

    return pipeline_spec.to_json()


def upload_environment(
    client: Client, repo: pfs.Repo, config: PpsConfig, script: bytes
) -> str:
    """Manages the pipeline's "environment" through a companion repo.

    The companion repo contains:
      - user_code.py     : the user's python script
      - requirements.txt : requirements optionally specified by the user
      - entrypoint.py    : the entrypoint for setting up and running the user code
      and the entrypoint script to orchestrate their code.

    Returns
    -------
        The branch name with the new commit as HEAD. We need to create a new branch
          as a workaround because specifying a commit within a pipeline's input
          spec is currently unsupported.
    """

    def entrypoint():
        """Entrypoint used by the PPS extension."""
        print("Greetings from the Pachyderm PPS Extension")

        from pathlib import Path
        from subprocess import run

        reqs = Path(__file__).parent.joinpath("requirements.txt")
        if reqs.exists():
            run(
                ["pip", "--disable-pip-version-check", "install", "-r", reqs.as_posix()]
            )
            print("Finished installing requirements")

        print("running user code")

        # gazelle:ignore user_code
        import user_code  # This runs the user's code.

    entrypoint_script = (
        f"{dedent(getsource(entrypoint))}\n\n"
        'if __name__ == "__main__":\n'
        "    entrypoint()\n"
    )

    master = pfs.Branch(repo=repo, name="master")
    with client.pfs.commit(branch=master) as commit:
        # Remove the old files
        commit.delete_file(path="/")

        # Upload the new files
        commit.put_file_from_bytes(path="/user_code.py", data=script)
        if config.requirements:
            with open(config.requirements, "rb") as reqs_file:
                commit.put_file_from_file(path="/requirements.txt", file=reqs_file)
        for external_file in config.external_files:
            with open(external_file, "rb") as external_file_data:
                commit.put_file_from_file(path=f'/{os.path.basename(external_file)}', file=external_file_data)
        commit.put_file_from_bytes(
            path="/entrypoint.py", data=entrypoint_script.encode("utf-8")
        )

    # Use the commit ID in the branch name to avoid name collisions.
    branch = pfs.Branch(repo=repo, name=f"commit_{commit.id}")
    client.pfs.create_branch(head=commit, branch=branch)

    return branch.name


class PPSClient:
    """Client interface for the PPS extension backend."""

    def __init__(self, client: Client, root_dir: Path):
        """
        client: The pachyderm client.
        root_dir: The root path from which all data and notebook files are relative.
        """
        self.nbconvert = PythonExporter()
        self.client = client
        self.root_dir = root_dir

    async def generate(self, path):
        """Generates the pipeline spec from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
        """
        get_logger().debug(f"path: {path}")

        path = self.root_dir.joinpath(path.lstrip("/")).resolve()
        if not path.exists():
            raise HTTPError(status_code=400, reason=f"notebook does not exist: {path}")

        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(status_code=400, reason=str(err))

        return create_pipeline_spec(config, "...")

    async def create(self, path: str, body: dict):
        """Creates the pipeline from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            body: The body of the request.
        """
        get_logger().debug(f"path: {path} | body: {body}")

        path = self.root_dir.joinpath(path.lstrip("/")).resolve()
        if not path.exists():
            raise HTTPError(status_code=400, reason=f"notebook does not exist: {path}")

        check = body.get("last_modified_time")
        if not check:
            raise HTTPError(
                status_code=400, reason="Bad Request: last_modified_time not specified"
            )
        check = datetime.fromisoformat(check.rstrip("Z"))

        last_modified = datetime.utcfromtimestamp(os.path.getmtime(path))
        if check != last_modified:
            raise HTTPError(
                status_code=400,
                reason=f"stale notebook: client time ({check}) != on-disk time ({last_modified})",
            )
        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(status_code=400, reason=str(err))

        if config.requirements and not os.path.exists(config.requirements):
            raise HTTPError(status_code=400, reason="requirements file does not exist")

        for external_file in config.external_files:
            if not os.path.exists(external_file):
                raise HTTPError(status_code=400, reason=f'external file {os.path.basename(external_file)} could not be found in the directory of the Jupyter notebook')

        script, _resources = self.nbconvert.from_filename(str(path))

        companion_repo = pfs.Repo(
            name=companion_repo_name(config),
            project=pfs.Project(name=config.pipeline.project.name),
        )
        for r in self.client.pfs.list_repo():
            same_project = config.pipeline.project.name == r.repo.project.name
            if (companion_repo.name == r.repo.name) and same_project:
                break
        else:
            self.client.pfs.create_repo(
                repo=companion_repo,
                description=f"files for running the {config.pipeline.name} pipeline",
            )

        companion_branch = upload_environment(
            self.client, companion_repo, config, script.encode("utf-8")
        )
        pipeline_spec = create_pipeline_spec(config, companion_branch)
        try:
            self.client.pps.create_pipeline_v2(
                create_pipeline_request_json=pipeline_spec, update=True, reprocess=True
            )
        except Exception as e:
            if hasattr(e, "details"):
                raise HTTPError(status_code=400, reason=e.details())
            raise HTTPError(
                status_code=500, reason=f"error creating pipeline: {repr(e)}"
            )

        return json.dumps(
            dict(
                message="Create pipeline request sent. You may monitor its status by running"
                ' "pachctl list pipelines" in a terminal.'
            )  # We can send back console link here.
        )
