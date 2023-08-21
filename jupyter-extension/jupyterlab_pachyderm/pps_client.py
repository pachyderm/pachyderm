import json
import yaml
import os.path
from dataclasses import dataclass, asdict
from datetime import datetime
from inspect import getsource
from pathlib import Path
from textwrap import dedent
from typing import Optional, Union

import python_pachyderm
from python_pachyderm.proto.v2.pps.pps_pb2 import Pipeline
from python_pachyderm.proto.v2.pfs.pfs_pb2 import Project
from nbconvert import PythonExporter
from tornado.web import HTTPError

from .log import get_logger

METADATA_KEY = 'pachyderm_pps'


class PPSClient:
    """Client interface for the PPS extension backend."""

    def __init__(self):
        self.nbconvert = PythonExporter()

    @staticmethod
    async def generate(path):
        """Generates the pipeline spec from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
        """
        get_logger().debug(f"path: {path}")

        path = Path(path.lstrip("/"))
        if not path.exists():
            raise HTTPError(
                status_code=400,
                reason=f"notebook does not exist: {path}"
            )

        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(status_code=400, reason=str(err))

        pipeline_spec = create_pipeline_spec(config, '...')
        return json.dumps(pipeline_spec)

    async def create(self, path: str, body: dict):
        """Creates the pipeline from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            body: The body of the request.
        """
        get_logger().debug(f"path: {path} | body: {body}")

        path = Path(path.lstrip("/"))
        if not path.exists():
            raise HTTPError(status_code=400, reason=f"notebook does not exist: {path}")

        check = body.get('last_modified_time')
        if not check:
            raise HTTPError(status_code=400,
                reason="Bad Request: last_modified_time not specified")
        check = datetime.fromisoformat(check.rstrip('Z'))

        last_modified = datetime.utcfromtimestamp(os.path.getmtime(path))
        if check != last_modified:
            raise HTTPError(
                status_code=400,
                reason=f"stale notebook: client time ({check}) != on-disk time ({last_modified})"
            )
        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(status_code=400, reason=str(err))

        if config.requirements and not os.path.exists(config.requirements):
            raise HTTPError(status_code=400, reason="requirements file does not exist")

        script, _resources = self.nbconvert.from_filename(str(path))

        client = python_pachyderm.Client()  # TODO: Auth?
        try:  # Verify connection to cluster.
            client.inspect_cluster()
        except Exception as e:
            raise HTTPError(
                status_code=500,
                reason=f"could not verify connection to Pachyderm cluster: {repr(e)}"
            )

        companion_repo = f"{config.pipeline.name}__context"
        for item in client.list_repo():
            same_project = config.pipeline.project.name == item.repo.project.name
            if (companion_repo == item.repo.name) and same_project:
                break
        else:
            client.create_repo(
                repo_name=companion_repo,
                project_name=config.pipeline.project.name,
                description=f"files for running the {config.pipeline.name} pipeline"
            )

        companion_branch = upload_environment(client, companion_repo, config, script.encode('utf-8'))
        pipeline_spec = create_pipeline_spec(config, companion_branch)
        try:
            client.create_pipeline_from_request(
                python_pachyderm.parse_dict_pipeline_spec(pipeline_spec)
            )
        except Exception as e:
            if hasattr(e, "details"):
                # Common case: pull error message out of Pachyderm RPC response
                raise HTTPError(status_code=400, reason=e.details())
            raise HTTPError(
                status_code=500,
                reason=f"error creating pipeline: {repr(e)}"
            )

        return json.dumps(
            dict(message="Create pipeline request sent. You may monitor its status by running"
                 " \"pachctl list pipelines\" in a terminal.")  # We can send back console link here.
        )


@dataclass
class PpsConfig:

    notebook_path: Path
    pipeline: Pipeline
    image: str
    requirements: Optional[str]
    port: str
    gpu_mode: str
    resource_spec: dict
    input_spec: dict  # We may be able to use the pachyderm SDK to parse and validate.

    @classmethod
    def from_notebook(cls, notebook_path: Union[str, Path]) -> "PpsConfig":
        """Parses a config from the metadata of a notebook file.

        Raises ValueError if required field is not specified.
        """
        notebook_path = Path(notebook_path)
        notebook_data = json.loads(notebook_path.read_bytes())

        metadata = notebook_data['metadata'].get(METADATA_KEY)
        if metadata is None:
            raise ValueError(f"{METADATA_KEY} not specified in notebook metadata")

        config = metadata.get('config')
        if config is None:
            raise ValueError(f"{METADATA_KEY}.config not specified in notebook metadata")

        pipeline_data = config.get('pipeline')
        if pipeline_data is None:
            raise ValueError("field pipeline not set")

        if 'project' in pipeline_data:
            project = Project(name=pipeline_data['project'].get('name') or 'default')
        else:
            project = Project(name='default')
        pipeline = Pipeline(name=pipeline_data.get('name'), project=project)

        image = config.get('image')
        if image is None:
            raise ValueError("field image not set")

        requirements = config.get('requirements')

        input_spec_str = config.get('input_spec')
        if input_spec_str is None:
            raise ValueError("field input_spec not set")
        input_spec = yaml.safe_load(input_spec_str)
        
        port = config.get('port')

        gpu_mode = config.get('gpu_mode')
        resource_spec_str = config.get('resource_spec')

        resource_spec = dict() if resource_spec_str is None else yaml.safe_load(resource_spec_str)

        return cls(
            notebook_path=notebook_path,
            pipeline=pipeline,
            image=image,
            requirements=requirements,
            input_spec=input_spec,
            port=port,
            gpu_mode=gpu_mode,
            resource_spec=resource_spec,
        )

    def to_dict(self):
        data = asdict(self)
        del data['notebook_path']
        return data


def create_pipeline_spec(config: PpsConfig, companion_branch: str) -> dict:
    companion_repo = f"{config.pipeline.name}__context"
    input_spec = dict(
        cross=[
            config.input_spec,
            dict(
                pfs=dict(
                    project=config.pipeline.project.name,
                    repo=companion_repo,
                    branch=companion_branch,
                    glob="/"
                )
            )
        ]
    )
    pipeline = dict(name=config.pipeline.name)
    if config.pipeline.project and config.pipeline.project.name:
        pipeline['project'] = dict(name=config.pipeline.project.name)
    pipelineSpec =  dict(
        pipeline=pipeline,
        description="Auto-generated from notebook",
        transform=dict(
            cmd=["python3", "-u", f"/pfs/{companion_repo}/entrypoint.py"],
            image=config.image
        ),

        input=input_spec,
        update=True,
        reprocess=True,
    )
    if config.port:
        pipelineSpec['service'] = dict(
            external_port=config.port,
            internal_port=config.port,
            type="LoadBalancer"
        )
    
    if config.gpu_mode == "Simple":
        pipelineSpec['resource_limits'] = dict(
            gpu=dict(
                type="nvidia.com/gpu",
                number="1",
            )
        )
    elif config.gpu_mode == "Advanced":
        pipelineSpec['resource_limits'] = config.resource_spec
        pass


    return pipelineSpec


def upload_environment(
        client: python_pachyderm.Client, repo: str, config: PpsConfig, script: bytes
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
            run(["pip", "--disable-pip-version-check", "install", "-r", reqs.as_posix()])
            print("Finished installing requirements")
        
        print("running user code")
        import user_code  # This runs the user's code.

    entrypoint_script = (
        f'{dedent(getsource(entrypoint))}\n\n'
        'if __name__ == "__main__":\n'
        '    entrypoint()\n'
    )
    project = config.pipeline.project.name
    with client.commit(repo, "master", project_name=project) as commit:
        # Remove the old files.
        for item in client.list_file(commit, "/"):
            client.delete_file(commit, item.file.path)

        # Upload the new files.
        client.put_file_bytes(commit, f"/user_code.py", script)
        if config.requirements:
            with open(config.requirements, "rb") as reqs_file:
                client.put_file_bytes(commit, "/requirements.txt", reqs_file)
        client.put_file_bytes(commit, "/entrypoint.py", entrypoint_script.encode('utf-8'))

    # Use the commit ID in the branch name to avoid name collisions.
    branch_name = f"commit_{commit.id}"
    client.create_branch(repo, branch_name, commit, project_name=project)

    return branch_name
