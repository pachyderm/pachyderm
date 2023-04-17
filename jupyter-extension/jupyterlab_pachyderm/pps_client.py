import json
import os.path
from dataclasses import dataclass
from datetime import datetime
from inspect import getsource
from pathlib import Path
from textwrap import dedent
from typing import Optional, Union

import python_pachyderm
from nbconvert import PythonExporter
from tornado.web import HTTPError

from .log import get_logger


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
            raise HTTPError(reason=f"notebook does not exist: {path}")

        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(reason=f"Bad Request: {err}")

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
            raise HTTPError(reason=f"notebook does not exist: {path}")

        check = body.get('last_modified_time')
        if not check:
            raise HTTPError(reason="Bad Request: last_modified_time not specified")
        check = datetime.fromisoformat(check.rstrip('Z'))

        last_modified = datetime.utcfromtimestamp(os.path.getmtime(path))
        get_logger().error(f"{check}, {last_modified}")
        if check != last_modified:
            raise HTTPError(
                reason=f"notebook verification failed: {check} != {last_modified}"
            )
        try:
            config = PpsConfig.from_notebook(path)
        except ValueError as err:
            raise HTTPError(reason=f"Bad Request: {err}")

        if config.requirements and not os.path.exists(config.requirements):
            raise HTTPError(reason="requirements file does not exist")

        script, _resources = self.nbconvert.from_filename(str(path))

        client = python_pachyderm.Client()  # TODO: Auth?
        companion_repo = f"{config.pipeline_name}__context"
        if companion_repo not in (item.repo.name for item in client.list_repo()):
            client.create_repo(
                repo_name=companion_repo,
                description=f"files for running the {config.pipeline_name} pipeline"
            )

        companion_branch = upload_environment(client, companion_repo, config, script.encode('utf-8'))
        pipeline_spec = create_pipeline_spec(config, companion_branch)
        client.create_pipeline_from_request(
            python_pachyderm.parse_dict_pipeline_spec(pipeline_spec)
        )

        return json.dumps(
            dict(message="Create pipeline request sent. You may monitor its status by running"
                 " \"pachctl list pipelines\" in a terminal.")  # We can send back console link here.
        )


@dataclass
class PpsConfig:

    notebook_path: Path
    pipeline_name: str
    image: str
    requirements: Optional[str]
    input_spec: dict  # We may be able to use the pachyderm SDK to parse and validate.

    @classmethod
    def from_notebook(cls, notebook_path: Union[str, Path]) -> "PpsConfig":
        """Parses a config from the metadata of a notebook file.

        Raises ValueError if required field is not specified.
        """
        notebook_path = Path(notebook_path)
        notebook_data = json.loads(notebook_path.read_bytes())

        config = notebook_data['metadata'].get('same_config')
        if config is None:
            raise ValueError("pps_config not specified in notebook metadata")

        pipeline_name = config['metadata'].get('name', None)
        if pipeline_name is None:
            raise ValueError("field pipeline_name not set")

        image = config['environments']['default'].get('image_tag', None)
        if image is None:
            raise ValueError("field image not set")

        requirements = config['notebook'].get('requirements')

        input_spec = config['run'].get('input', None)
        if input_spec is None:
            raise ValueError("field input_spec not set")
        input_spec = json.loads(input_spec)

        return cls(
            notebook_path=notebook_path,
            pipeline_name=pipeline_name,
            image=image,
            requirements=requirements,
            input_spec=input_spec
        )


def create_pipeline_spec(config: PpsConfig, companion_branch: str) -> dict:
    companion_repo = f"{config.pipeline_name}__context"
    input_spec = dict(
        cross=[
            config.input_spec,
            {'pfs': dict(repo=companion_repo, branch=companion_branch, glob="/")}
        ]
    )
    return dict(
        pipeline=dict(name=config.pipeline_name),
        description="Auto-generated from notebook",
        transform=dict(
            cmd=["python3", f"/pfs/{companion_repo}/entrypoint.py"],
            image=config.image
        ),
        input=input_spec,
        update=True,
        reprocess=True,
    )


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

        reqs = Path("requirements.txt")
        if reqs.exists():
            run(["pip", "--disable-pip-version-check", "install", "-r", reqs.as_posix()])

        import user_code  # This runs the user's code.

    entrypoint_script = (
        f'{dedent(getsource(entrypoint))}\n\n'
        'if __name__ == "__main__":\n'
        '    entrypoint()\n'
    )

    with client.commit(repo, "master") as commit:
        # Remove the old files.
        for item in client.list_file((repo, "master"), "/"):
            client.delete_file(commit, item.file.path)

        # Upload the new files.
        client.put_file_bytes(commit, f"/user_code.py", script)
        if config.requirements:
            with open(config.requirements, "rb") as reqs_file:
                client.put_file_bytes(commit, "/requirements.txt", reqs_file)
        client.put_file_bytes(commit, "/entrypoint.py", entrypoint_script.encode('utf-8'))

    # Use the commit ID in the branch name to avoid name collisions.
    branch_name = f"commit_{commit.id}"
    client.create_branch(repo, branch_name, commit)

    return branch_name
