import random
from typing import Tuple

from pachyderm_sdk.client import Client
from pachyderm_sdk.api import pfs, pps


class TestClient(Client):
    """This is a test client that keeps track of the resources created and
    cleans them up once the test is complete.

    TODO:
        * Add resource names when using verbosity
    """

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
        if self._project_exists(project):
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
                cmd=["sh"],
                image="alpine",
                stdin=[f"cp /pfs/{repo.name}/*.dat /pfs/out/"]
            )
        )
        self.pipelines.append(pipeline)

        with self.pfs.commit(branch=pfs.Branch(repo=repo, name="master")) as commit:
            commit.put_file_from_bytes(path="file.dat", data=b"DATA")
        commit.wait()

        pipeline_info = self.pps.inspect_pipeline(pipeline=pipeline, details=True)
        job_info = next(self.pps.list_job(pipeline=pipeline))
        return pipeline_info, job_info

    def tear_down(self):
        for pipeline in self.pipelines:
            self.pps.delete_pipeline(pipeline=pipeline, force=True)
        for repo in self.repos:
            self.pfs.delete_repo(repo=repo, force=True)
        for project in self.projects:
            if self._project_exists(project):
                self.pfs.delete_project(project=project, force=True)

    def _project_exists(self, project: pfs.Project) -> bool:
        all_projects_names = {info.project.name for info in self.pfs.list_project()}
        return project.name in all_projects_names

    def _generate_name(self) -> str:
        name: str = (
            self.id
                .replace("/", "-")
                .replace(":", "-")
                .replace(".py", "")
        )[:40]  # TODO: Make this the maximum it can be.
        name = name[:name.find("[")]
        return name
