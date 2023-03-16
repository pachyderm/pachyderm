import os.path
import json
import subprocess
import sys
from pathlib import Path
from tempfile import NamedTemporaryFile

from tornado.httpclient import AsyncHTTPClient

from .log import get_logger


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

    async def create(self, path, config):
        """Creates the pipeline from the Notebook file specified.

        Args:
            path: The path (within Jupyter) to the Notebook file.
            config: The PPS configuration for the Notebook file.
        """
        get_logger().debug(f"path: {path} | body: {config}")
        input_spec = config.get("input")
        pipeline_name = config.get("pipeline_name")

        with NamedTemporaryFile() as temp_config:
            Path(temp_config.name).write_text(
                "apiVersion: sameproject.ml/v1alpha1\n"
                "metadata:\n"
                f"  name: {pipeline_name}\n"
                f"  version: 0.0.1\n"
                "environments:\n"
                "  default:\n"
                "    image_tag: combinatorml/jupyterlab-tensorflow-opencv:0.9\n"
                "notebook:\n"
                f"  name: TestNotebook\n"
                f"  path: {Path(__file__).parent}/tests/data/TestNotebook.ipynb\n"
                # f"  name: {os.path.basename(path)}\n"
                # f"  path: {path}\n"
            )

            call = subprocess.run(
                [sys.executable, "-m", "sameproject.main", "run", "--same-file", temp_config.name, "--target", "pachyderm", "--input", json.dumps(input_spec)],
                check=True
            )
        return json.dumps(dict())


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
