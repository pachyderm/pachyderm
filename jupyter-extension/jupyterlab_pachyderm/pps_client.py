import os.path
import json
from dataclasses import dataclass
from inspect import getsource
from textwrap import dedent
from typing import Optional

import python_pachyderm
from nbconvert import PythonExporter
from tornado.web import HTTPError

from .log import get_logger


class PPSClient:
    """Client interface for the PPS extension backend."""

    def __init__(self):
        self.nbconvert = PythonExporter()

    async def generate(self, path, config):
        """Generates the pipeline spec from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            config: The PPS configuration for the Notebook file.
        """
        get_logger().debug(f"path: {path} | body: {config}")
        config = PpsConfig.from_json(config)
        pipeline_spec = create_pipeline_spec(config, '...')
        return json.dumps(pipeline_spec)

    async def create(self, path: str, config: dict):
        """Creates the pipeline from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            config: The PPS configuration for the Notebook file.
        """
        get_logger().debug(f"path: {path} | body: {config}")

        config = PpsConfig.from_json(config)
        path = path.lstrip("/")
        if not os.path.exists(path):
            raise HTTPError(reason=f"notebook does not exist: {path}")
        if config.requirements and not os.path.exists(config.requirements):
            raise HTTPError(reason="requirements file does not exist")

        script, _resources = self.nbconvert.from_filename(path)

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
            dict(message=None)  # We can send back console link here.
        )


@dataclass
class PpsConfig:

    pipeline_name: str
    image: str
    requirements: Optional[str]
    input_spec: dict  # We may be able to use the pachyderm SDK to parse and validate.

    @classmethod
    def from_json(cls, config) -> "PpsConfig":
        """Parses a config from a json object.

        Raises HTTPError with 500 code if required field is not specified.
        """
        pipeline_name = config.get('pipeline_name', None)
        if pipeline_name is None:
            raise HTTPError(reason=f"Bad Request: field pipeline_name not set")

        image = config.get('image', None)
        if image is None:
            raise HTTPError(reason=f"Bad Request: field image not set")

        requirements = config.get('requirements')

        input_spec = config.get('input_spec', None)
        if input_spec is None:
            raise HTTPError(reason=f"Bad Request: field input_spec not set")

        return cls(
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
