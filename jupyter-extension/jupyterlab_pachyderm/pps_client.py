import os.path
import json
import subprocess
import sys
from pathlib import Path
from tempfile import NamedTemporaryFile

from tornado.httpclient import AsyncHTTPClient
from tornado.web import HTTPError

from .log import get_logger

SAME_METADATA_FIELDS = ('apiVersion', 'environments', 'metadata', 'run', 'notebook')


class PPSClient:
    """Client interface for the PPS extension backend."""

    def __init__(self):
        self.client = AsyncHTTPClient()

    async def generate(self, path, config):
        """Generates the pipeline spec from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            config: The PPS configuration for the Notebook file.
        """
        get_logger().debug(f"path: {path} | body: {config}")
        # script = await self.client.fetch(f"http://localhost:8888/nbconvert/python{path}")
        script_name = os.path.basename(path)
        pipeline_spec = create_pipeline_spec("test_pipeline", "python:3.10", dict(pfs=dict(repo="data")), script_name)
        return json.dumps(pipeline_spec)

    async def create(self, path: str, config: dict):
        """Creates the pipeline from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            config: The PPS configuration for the Notebook file.
        """
        get_logger().debug(f"path: {path} | body: {config}")

        # Validate config structure.
        # Error codes are 500, we are unable to override this.
        for field in SAME_METADATA_FIELDS:
            if field not in config:
                raise HTTPError(reason=f"Bad Request: field {field} not set")
        if 'input' not in config['run']:
            raise HTTPError(reason=f"Bad Request: field run.input not set")

        input_spec = config['run'].pop("input")  # Input not allowed in SAME metadata.
        # Paths are relative to the config file. Since the config is being written into
        #  a temporary file, we need to specify an absolute path.
        config['notebook']['path'] = os.path.join(os.getcwd(), os.path.relpath(path, '/'))
        config['notebook']['name'] = os.path.basename(path)

        if 'requirements' in config['notebook'] and not config['notebook']['requirements']:
            # Remove requirements field if none are specified.
            del config['notebook']['requirements']

        with NamedTemporaryFile() as temp_config:
            Path(temp_config.name).write_text(json.dumps(config))
            command = [sys.executable, "-m", "sameproject.main", "run"]
            command.extend(["--same-file", temp_config.name])
            command.extend(["--target", "pachyderm"])
            command.extend(["--input", input_spec])
            call = subprocess.run(command, capture_output=True)
            if call.returncode:
                raise HTTPError(
                    code=500,
                    log_message=call.stderr.decode(),
                    reason="Internal Error: SAME could not create pipeline"
                )

        return json.dumps(
            dict(message=None)  # We can send back console link here.
        )


def create_pipeline_spec(
    pipeline_name: str,
    image: str,
    input_spec: dict,
    script_name: str,
) -> dict:
    """Generates the pipelines spec. Currently copied from """
    companion_repo = f"{pipeline_name}__context"
    micro_entrypoint = (
        'print("Greetings from the Pachyderm PPS Extension"); '
        + 'import sys; '
        + 'from importlib import import_module; '
        + 'from pathlib import Path; '
        + 'from subprocess import run; '

        + 'root_module = sys.argv[1]; '
        + f'context_dir = Path("/pfs", "{companion_repo}"); '
        + 'reqs = context_dir / "requirements.txt"; '
        + 'reqs.exists() and run(["pip", "--disable-pip-version-check", "install", "-r", reqs.as_posix()]); '
        + 'sys.path.append(context_dir.as_posix()); '
        + 'script = import_module(root_module); '
        + 'script.root()'
    )
    cmd = ["python3", "-c", micro_entrypoint, script_name]
    return dict(
        pipeline=dict(name=pipeline_name),
        description="Auto-generated from notebook",
        transform=dict(cmd=cmd, image=image),
        input=input_spec,
        update=True,
        reprocess=True
    )
