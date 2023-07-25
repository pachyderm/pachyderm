"""Tests for PPS-related functionality."""
import grpc

from tests.fixtures import *
from tests.utils import count

from pachyderm_sdk.api import pps


class TestUnitPipeline:
    """Unit tests for the pipeline management API."""

    @staticmethod
    def test_inspect_pipeline(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline = pipeline_info.pipeline

        response = client.pps.inspect_pipeline(pipeline=pipeline)
        assert response.pipeline.name == pipeline.name

        response = client.pps.inspect_pipeline(pipeline=pipeline, history=-1)
        assert response.pipeline.name == pipeline.name

    @staticmethod
    def test_list_pipeline(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        response = list(client.pps.list_pipeline())
        assert pipeline_info.pipeline.name in [p.pipeline.name for p in response]

    @staticmethod
    def test_delete_pipeline(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        client.pps.delete_pipeline(pipeline=pipeline_info.pipeline)
        assert not client.pps.pipeline_exists(pipeline_info.pipeline)

    @staticmethod
    def test_restart_pipeline(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline = pipeline_info.pipeline

        client.pps.stop_pipeline(pipeline=pipeline)
        response = client.pps.inspect_pipeline(pipeline=pipeline)
        assert response.stopped

        client.pps.start_pipeline(pipeline=pipeline)
        response = client.pps.inspect_pipeline(pipeline=pipeline)
        assert not response.stopped


class TestUnitJob:
    """Unit tests for the job management API."""

    @staticmethod
    def test_list_subjob(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline = pipeline_info.pipeline

        jobs = client.pps.list_job(pipeline=pipeline)
        assert count(jobs) >= 1

        jobs = client.pps.list_job(pipeline=pipeline, projects=[pipeline.project])
        assert count(jobs) >= 1

        jobs = client.pps.list_job(
            pipeline=pipeline,
            projects=[pipeline.project],
            input_commit=pipeline_info.spec_commit,
        )
        assert count(jobs) >= 1

    @staticmethod
    def test_list_job(client: TestClient, default_project: bool):
        _pipeline_info_1, _job_info_1 = client.new_pipeline(default_project)
        _pipeline_info_2, _job_info_2 = client.new_pipeline(default_project)

        jobs = client.pps.list_job()
        assert count(jobs) >= 4

    @staticmethod
    def test_inspect_job(client: TestClient, default_project: bool):
        pipeline_info_1, job_info_1 = client.new_pipeline(default_project)
        job_info = client.pps.inspect_job(job=job_info_1.job)
        assert job_info.job.id == job_info_1.job.id

    @staticmethod
    def test_stop_job(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        client.pps.stop_job(job=job_info.job)

        job_info = client.pps.inspect_job(job=job_info.job, wait=True)

        # We race to stop the job before it finishes.
        # If we lose the race, it will be in state JOB_SUCCESS.
        assert job_info.state in [
            pps.JobState.JOB_KILLED,
            pps.JobState.JOB_SUCCESS,
        ]

    @staticmethod
    def test_delete_job(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        client.pps.delete_job(job=job_info.job)
        assert not client.pps.job_exists(job=job_info.job)


class TestUnitDatums:
    """Unit tests for the datum management API."""

    @staticmethod
    def test_datums(client: TestClient, default_project: bool):
        # TODO: Why is this test sometimes slow?
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline, job = pipeline_info.pipeline, job_info.job

        client.pps.inspect_job(job=job, wait=True)
        datums = list(client.pps.list_datum(job=job))
        assert len(datums) == 1
        datum = client.pps.inspect_datum(datum=datums[0].datum)
        assert datum.state == pps.DatumState.SUCCESS

        error_match = (
            rf"datum matching filter \[.*\] could not be found for job ID {job_info.job.id}"
        )
        with pytest.raises(grpc.RpcError, match=error_match):
            client.pps.restart_datum(job=job)


class TestUnitLogs:
    """Unit tests for the log retrieval API."""

    @staticmethod
    def test_get_pipeline_logs(client: TestClient):
        pipeline_info, job_info = client.new_pipeline()
        pipeline, job = pipeline_info.pipeline, job_info.job

        # Just make sure these spit out some logs
        logs = client.pps.get_logs(pipeline=pipeline, follow=True)
        assert next(logs) is not None
        del logs

        logs = client.pps.get_logs(pipeline=pipeline, follow=True, master=True)
        assert next(logs) is not None
        del logs

    @staticmethod
    def test_get_job_logs(client: TestClient):
        pipeline_info, job_info = client.new_pipeline()
        pipeline, job = pipeline_info.pipeline, job_info.job

        # Just make sure these spit out some logs
        logs = client.pps.get_logs(job=job, follow=True)
        assert next(logs) is not None
        del logs


class TestUnitMisc:
    """Unit tests for miscellaneous API."""

    @staticmethod
    def test_run_cron(client: TestClient, default_project: bool):
        pipeline_info, job_info = client.new_pipeline(default_project)
        pipeline, job = pipeline_info.pipeline, job_info.job

        # this should trigger an error because the sandbox pipeline
        # does not have a cron input
        with pytest.raises(grpc.RpcError, match=r"pipeline.*have a cron input"):
            client.pps.run_cron(pipeline=pipeline)

    @staticmethod
    def test_secrets(client: TestClient):
        secret_name = "bogus-secret"
        secret = pps.Secret(name=secret_name)
        try:
            client.pps.create_secret(name=secret_name, data={"super": "secret"})
            secret_info = client.pps.inspect_secret(secret=secret)
            assert secret_info.secret.name == secret_name

            response = client.pps.list_secret()
            secrets = response.secret_info
            assert len(secrets) == 1
            assert secrets[0].secret.name == secret_name
            client.pps.delete_secret(secret=secret)

        except Exception as err:
            client.pps.delete_secret(secret=secret)
            raise err

        with pytest.raises(grpc.RpcError):
            client.pps.inspect_secret(secret=secret)

        response = client.pps.list_secret()
        assert len(response.secret_info) == 0
