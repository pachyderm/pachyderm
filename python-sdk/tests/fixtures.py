import os
import random
from typing import Tuple

import pytest

from pachyderm_sdk.api import pfs, pps
from pachyderm_sdk.client import Client as _Client
from pachyderm_sdk.constants import AUTH_TOKEN_ENV


@pytest.fixture(params=[True, False])
def default_project(request) -> bool:
    """Parametrized fixture with values True|False.

    Use this fixture to easily test resources against
      default and non-default projects.
    """
    return request.param


@pytest.fixture
def client(request) -> "TestClient":
    client = TestClient(
        nodeid=request.node.nodeid,
        host=os.environ.get("PACH_PYTHON_TEST_HOST"),
        port=os.environ.get("PACH_PYTHON_TEST_PORT"),
    )
    yield client
    client.tear_down()


@pytest.fixture
def auth_client(request) -> "TestClient":
    client = TestClient(
        nodeid=request.node.nodeid,
        host=os.environ.get("PACH_PYTHON_TEST_HOST"),
        port=os.environ.get("PACH_PYTHON_TEST_PORT_ENTERPRISE"),
        auth_token=os.environ.get(AUTH_TOKEN_ENV),
    )
    yield client
    client.tear_down()


class TestClient(_Client):
    """This is a test client that keeps track of the resources created and
    cleans them up once the test is complete.

    TODO:
        * Add resource names when using verbosity
    """

    __test__ = False

    def __init__(self, *args, nodeid: str, **kwargs):
        """
        Args:
            nodeid: The pytest nodeid used to label resources (in their descriptions)
        """
        super().__init__(*args, **kwargs)
        self.id = nodeid
        self.projects = []
        self.repos = []
        self.pipelines = []

    def new_project(self) -> pfs.Project:
        project = pfs.Project(name=f"proj{random.randint(100, 999)}")
        if self.pfs.project_exists(project):
            self.pfs.delete_project(project=project, force=True)
        self.pfs.create_project(project=project, description=self.id)
        self.projects.append(project)
        return project

    def new_repo(self, default_project: bool = True) -> pfs.Repo:
        if not default_project:
            project = self.new_project()
        else:
            # By having the project name be an empty string we check
            #   that the server is properly defaulting the value in
            #   client requests.
            project = pfs.Project(name="")

        repo = pfs.Repo(name=self._generate_name(), type="user", project=project)
        self.pfs.delete_repo(repo=repo, force=True)
        self.pfs.create_repo(repo=repo, description=self.id)
        self.pfs.create_branch(branch=pfs.Branch.from_uri(f"{repo}@master"))
        self.repos.append(repo)
        return repo

    def new_pipeline(
        self, default_project: bool = True
    ) -> Tuple[pps.PipelineInfo, pps.JobInfo]:
        repo = self.new_repo(default_project)
        pipeline = pps.Pipeline(project=repo.project, name=self._generate_name())
        self.pps.delete_pipeline(pipeline=pipeline, force=True)
        self.pps.create_pipeline(
            pipeline=pipeline,
            input=pps.Input(pfs=pps.PfsInput(glob="/*", repo=repo.name)),
            transform=pps.Transform(
                cmd=["sh"], image="alpine", stdin=[f"cp /pfs/{repo.name}/*.dat /pfs/out/"]
            ),
        )
        self.pipelines.append(pipeline)

        with self.pfs.commit(branch=pfs.Branch(repo=repo, name="master")) as commit:
            commit.put_file_from_bytes(path="file.dat", data=b"DATA")
        commit.wait()

        pipeline_info = self.pps.inspect_pipeline(pipeline=pipeline, details=True)
        job_info = next(self.pps.list_job(pipeline=pipeline))
        return pipeline_info, job_info

    def tear_down(self):
        self.transaction_id = None
        for pipeline in self.pipelines:
            self.pps.delete_pipeline(pipeline=pipeline, force=True)
        for repo in self.repos:
            self.pfs.delete_repo(repo=repo, force=True)
        for project in self.projects:
            if self.pfs.project_exists(project):
                self.pfs.delete_project(project=project, force=True)

    def _generate_name(self) -> str:
        # fmt: off
        name: str = (
            self.id
                .replace("/", "-")
                .replace(":", "-")
                .replace(".py", "")
        )[:40]  # TODO: Make this the maximum it can be.
        # fmt: on
        name = f"{name[:name.find('[')]}-{random.randint(100, 999)}"
        return name
