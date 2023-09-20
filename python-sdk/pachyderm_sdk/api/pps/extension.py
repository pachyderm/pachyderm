"""Handwritten classes/methods that augment the existing PPS API."""
import base64
import json
from typing import Dict

import grpc
from betterproto.lib.google.protobuf import Empty

from . import ApiStub as _GeneratedApiStub
from . import (
    Job,
    Pipeline,
    PipelineInfo,
)


class ApiStub(_GeneratedApiStub):
    def inspect_pipeline(
        self,
        *,
        pipeline: "Pipeline" = None,
        details: bool = False,
        history: int = 0,
    ) -> "PipelineInfo":
        """Inspects a pipeline.

        Parameters
        ----------
        pipeline : pps.Pipeline
            The pipeline to inspect.
        details : bool, optional
            If true, return pipeline details.
        history : int, optional
            Indicates to return historical versions of `pipeline_name`.
            Semantics are:

            - 0: Return current version of `pipeline_name`
            - 1: Return the above and `pipeline_name` from the next most recent version.
            - 2: etc.
            - -1: Return all historical versions of `pipeline_name`.

        Returns
        -------
        pps.PipelineInfo

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pps
        >>> client: Client
        >>> pipeline_info = client.pps.inspect_pipeline(
        >>>     pipeline=pps.Pipeline(name="foo")
        >>> )
        """
        if history:
            response = self.list_pipeline(
                pipeline=pipeline, history=history, details=details
            )
            try:
                return next(response)
            except StopIteration:
                raise ValueError("invalid pipeline")
        return super().inspect_pipeline(pipeline=pipeline, details=details)

    def pipeline_exists(self, pipeline: "Pipeline") -> bool:
        """Checks whether a pipeline exists.

        Parameters
        ----------
        pipeline: pps.Pipeline
            The pipeline to check.

        Returns
        -------
        bool
            Whether the pipeline exists.
        """
        try:
            super().inspect_pipeline(pipeline=pipeline)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    def job_exists(self, job: "Job") -> bool:
        """Checks whether a job exists.

        Parameters
        ----------
        job: pps.Job
            The job to check.

        Returns
        -------
        bool
            Whether the job exists.
        """
        try:
            super().inspect_job(job=job)
            return True
        except grpc.RpcError as err:
            err: grpc.Call
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err

    # noinspection PyMethodOverriding
    def create_secret(
        self,
        *,
        name: str,
        data: Dict,
        labels: Dict[str, str] = None,
        annotations: Dict[str, str] = None,
    ) -> Empty:
        """Creates a new secret.

        Parameters
        ----------
        name : str
            The name of the secret.
        data : Dict[str, Union[str, bytes]]
            The data to store in the secret. Each key must consist of
            alphanumeric characters ``-``, ``_`` or ``.``.
        labels : Dict[str, str], optional
            Kubernetes labels to attach to the secret.
        annotations : Dict[str, str], optional
            Kubernetes annotations to attach to the secret.
        """
        encoded_data = {}
        for k, v in data.items():
            if isinstance(v, str):
                v = v.encode("utf8")
            encoded_data[k] = base64.b64encode(v).decode("utf8")

        file = json.dumps(
            {
                "kind": "Secret",
                "apiVersion": "v1",
                "metadata": {
                    "name": name,
                    "labels": labels,
                    "annotations": annotations,
                },
                "data": encoded_data,
            }
        ).encode()

        return super().create_secret(file=file)
