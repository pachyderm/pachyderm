"""Tests for PPS-related functionality."""

import time

import betterproto
import grpc

from tests.test_client import TestClient
from tests.fixtures import *
from tests.utils import count

from pachyderm_sdk.api import pps


class Sandbox:
    def __init__(self, test_name):
        client = python_pachyderm.experimental.Client()
        client.delete_all()
        (
            commit,
            input_repo_name,
            pipeline_repo_name,
            project_name,
        ) = util.create_test_pipeline(client, test_name)
        self.client = client
        self.commit = commit
        self.input_repo_name = input_repo_name
        self.pipeline_repo_name = pipeline_repo_name
        self.project_name = project_name

    def wait(self):
        return self.client.wait_commit(self.commit.id)[0].commit.id


class TestUnitPipeline:
    """Unit tests for the pipeline management API."""

    @staticmethod
    def test_inspect_pipeline(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline = pipeline_info.pipeline

        response = client.pps.inspect_pipeline(pipeline=pipeline)
        assert pipeline.pipeline.name == sandbox.pipeline_repo_name
        pipelines = list(
            sandbox.client.inspect_pipeline(
                sandbox.pipeline_repo_name, project_name=sandbox.project_name, history=-1
            )
        )
        assert sandbox.pipeline_repo_name in [p.pipeline.name for p in pipelines]

    def test_list_pipeline():
        sandbox = Sandbox("list_pipeline")
        pipelines = list(sandbox.client.list_pipeline())
        assert sandbox.pipeline_repo_name in [p.pipeline.name for p in pipelines]
        pipelines = list(sandbox.client.list_pipeline(history=-1))
        assert sandbox.pipeline_repo_name in [p.pipeline.name for p in pipelines]

    def test_delete_pipeline():
        sandbox = Sandbox("delete_pipeline")
        orig_pipeline_count = len(list(sandbox.client.list_pipeline()))
        time.sleep(1)
        sandbox.client.delete_pipeline(
            sandbox.pipeline_repo_name, project_name=sandbox.project_name
        )
        assert len(list(sandbox.client.list_pipeline())) == orig_pipeline_count - 1


class TestUnitJob:
    """Unit tests for the job management API."""

    @staticmethod
    def test_list_subjob(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline = pipeline_info.pipeline

        jobs = client.pps.list_job(pipeline=pipeline)
        assert count(jobs) >= 1

        jobs = client.pps.list_job(
            pipeline=pipeline,
            projects=[pipeline.project]
        )
        assert count(jobs) >= 1

        jobs = client.pps.list_job(
            pipeline=pipeline,
            projects=[pipeline.project],
            input_commit=pipeline_info.spec_commit,  # TODO: What is spec commit?
        )
        assert count(jobs) >= 1

    @staticmethod
    def test_list_job(client: TestClient, default_project: bool):
        _pipeline_info_1, _job_info_1 = client.new_pipeline(default_project)
        _pipeline_info_2, _job_info_2 = client.new_pipeline(default_project)

        jobs = client.pps.list_job()
        assert count(jobs) == 4

    @staticmethod
    def test_inspect_job(client: TestClient, default_project: bool):
        pipeline_info_1, job_info_1 = client.new_pipeline(default_project)
        _pipeline_info_2, _job_info_2 = client.new_pipeline(default_project)

        job_info = client.pps.inspect_job(job=job_info_1.job)

        assert job_info.job.id == pipeline_info_1.spec_commit.id
        assert count(client.pps.list_job()) == 4

    @staticmethod
    def test_stop_job(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        client.pps.stop_job(job=job_info.job)

        # This is necessary because `StopJob` does not wait for the job to be
        # killed before returning a result.
        # TODO: remove once this is fixed:
        # https://github.com/pachyderm/pachyderm/issues/3856
        # TODO: Can we just use wait?
        time.sleep(1)
        job_info = client.pps.inspect_job(job=job_info.job, wait=False)

        # We race to stop the job before it finishes - if we lose the race, it will
        # be in state JOB_SUCCESS
        assert job_info.state in [
            pps.JobState.JOB_KILLED,
            pps.JobState.JOB_SUCCESS,
        ]

    @staticmethod
    def test_delete_job(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        orig_job_count = count(client.pps.list_job())
        client.pps.delete_job(job=job_info.job)
        assert count(client.pps.list_job()) == orig_job_count - 1


def test_datums():
    sandbox = Sandbox("datums")
    pipeline_name = sandbox.pipeline_repo_name
    job_id = sandbox.wait()

    # flush the job so it fully finishes
    list(sandbox.client.wait_commit(sandbox.commit.id))

    datums = list(
        sandbox.client.list_datum(
            pipeline_name, job_id, project_name=sandbox.project_name
        )
    )
    assert len(datums) == 1
    datum = sandbox.client.inspect_datum(
        pipeline_name, job_id, datums[0].datum.id, project_name=sandbox.project_name
    )
    assert datum.state == pps_proto.DatumState.SUCCESS

    with pytest.raises(
        python_pachyderm.RpcError,
        match=r"datum matching filter \[.*\] could not be found for job ID {}".format(
            job_id
        ),
    ):
        sandbox.client.restart_datum(
            pipeline_name, job_id, project_name=sandbox.project_name
        )


def test_restart_pipeline():
    sandbox = Sandbox("restart_job")

    sandbox.client.stop_pipeline(
        sandbox.pipeline_repo_name, project_name=sandbox.project_name
    )
    pipeline = list(
        sandbox.client.inspect_pipeline(
            sandbox.pipeline_repo_name, project_name=sandbox.project_name
        )
    )[0]
    assert pipeline.stopped

    sandbox.client.start_pipeline(
        sandbox.pipeline_repo_name, project_name=sandbox.project_name
    )
    pipeline = list(
        sandbox.client.inspect_pipeline(
            sandbox.pipeline_repo_name, project_name=sandbox.project_name
        )
    )[0]
    assert not pipeline.stopped


def test_run_cron():
    sandbox = Sandbox("run_cron")

    # flush the job so it fully finishes
    list(sandbox.client.wait_commit(sandbox.commit.id))

    # this should trigger an error because the sandbox pipeline doesn't have a
    # cron input
    # NOTE: `e` is used after the context
    with pytest.raises(python_pachyderm.RpcError, match=r"pipeline.*have a cron input"):
        sandbox.client.run_cron(
            sandbox.pipeline_repo_name, project_name=sandbox.project_name
        )


def test_secrets():
    client = python_pachyderm.experimental.Client()
    secret_name = util.test_repo_name("test-secrets")

    client.create_secret(
        secret_name,
        {
            "mykey": "my-value",
        },
    )

    secret = client.inspect_secret(secret_name)
    assert secret.secret.name == secret_name

    secrets = client.list_secret()
    assert len(secrets) == 1
    assert secrets[0].secret.name == secret_name

    client.delete_secret(secret_name)

    with pytest.raises(python_pachyderm.RpcError):
        client.inspect_secret(secret_name)

    secrets = client.list_secret()
    assert len(secrets) == 0


def test_get_pipeline_logs():
    sandbox = Sandbox("pipeline_logs")
    sandbox.wait()

    # Just make sure these spit out some logs
    logs = sandbox.client.get_pipeline_logs(
        sandbox.pipeline_repo_name, project_name=sandbox.project_name, follow=True
    )
    assert next(logs) is not None
    del logs

    logs = sandbox.client.get_pipeline_logs(
        sandbox.pipeline_repo_name,
        project_name=sandbox.project_name,
        master=True,
        follow=True,
    )
    assert next(logs) is not None
    del logs


def test_get_job_logs():
    sandbox = Sandbox("get_job_logs")
    job_id = sandbox.wait()
    pipeline_name = sandbox.pipeline_repo_name

    # Wait for the job to complete
    commit = python_pachyderm.experimental.pfs.Commit(
        repo=pipeline_name, id=job_id, project=sandbox.project_name
    )
    sandbox.client.wait_commit(commit)

    # Just make sure these spit out some logs
    logs = sandbox.client.get_job_logs(
        pipeline_name, job_id, project_name=sandbox.project_name, follow=True
    )
    assert next(logs) is not None
    del logs


def test_create_pipeline():
    client = python_pachyderm.experimental.Client()
    client.delete_all()

    input_repo_name = util.create_test_repo(client, "input_repo_test_create_pipeline")

    client.create_pipeline(
        "pipeline_test_create_pipeline",
        transform=pps_proto.Transform(
            cmd=["sh"],
            image="alpine",
            stdin=["cp /pfs/{}/*.dat /pfs/out/".format(input_repo_name)],
        ),
        input=pps_proto.Input(pfs=pps_proto.PfsInput(glob="/*", repo=input_repo_name)),
    )
    assert len(list(client.list_pipeline())) == 1


def test_create_pipeline_from_request():
    client = python_pachyderm.experimental.Client()

    repo_name = util.create_test_repo(client, "test_create_pipeline_from_request")
    pipeline_name = util.test_repo_name("test_create_pipeline_from_request")

    # more or less a copy of the opencv demo's edges pipeline spec
    client.create_pipeline_from_request(
        pps_proto.CreatePipelineRequest(
            pipeline=pps_proto.Pipeline(name=pipeline_name),
            description="A pipeline that performs image edge detection by using the OpenCV library.",
            input=pps_proto.Input(
                pfs=pps_proto.PfsInput(
                    glob="/*",
                    repo=repo_name,
                ),
            ),
            transform=pps_proto.Transform(
                cmd=["echo", "hi"],
                image="pachyderm/opencv",
            ),
        )
    )

    assert any(p.pipeline.name == pipeline_name for p in list(client.list_pipeline()))
