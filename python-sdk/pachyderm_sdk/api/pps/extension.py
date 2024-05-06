"""Handwritten classes/methods that augment the existing PPS API."""

import base64
from collections.abc import Callable
import json
from queue import SimpleQueue
from typing import Dict, Iterable, Iterator, List

import grpc
from betterproto.lib.google.protobuf import Empty
from pachyderm_sdk.api.pps import CreateDatumRequest, DatumInfo

from . import ApiStub as _GeneratedApiStub
from . import (
    ContinueCreateDatumRequest,
    CreateDatumRequest,
    Input,
    Job,
    Pipeline,
    PipelineInfo,
    StartCreateDatumRequest,
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
            response = self.list_pipeline(pipeline=pipeline, history=history)
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

    class CreateDatumClient:
        """Client for creating datums. Note that this class shouldn't be
        instantiated directly. Instead, use the `create_datum` method to return
        a new instance.


        Methods
        -------
        first_batch(input, number) -> List[DatumInfo]:
            Returns the first batch of datums for the provided input spec.
        next_batch(number) -> List[DatumInfo]:
            Returns subsequent batches of datums. `first_batch` must be called first.

        Examples
        --------
        >>> from pachyderm_sdk import Client
        >>> from pachyderm_sdk.api import pps
        >>>
        >>> client = Client()
        >>> cdc = client.pps.create_datum()
        >>> input = pps.Input(pfs=pps.PFSInput(repo="repo", glob="/*"))
        >>>
        >>> datums = cdc.first_batch(input, number=10)
        >>> for datum in datums:
        >>>     print(datum)
        >>>
        >>> more_datums = cdc.next_batch(number=10)
        >>> for datum in more_datums:
        >>>     print(datum)
        """

        def __init__(
            self,
            create_datum: Callable[[Iterable["CreateDatumRequest"]], Iterator["DatumInfo"]],
        ):
            self.queue = SimpleQueue()
            self.response_it = create_datum(self)
            self.done = False

        def first_batch(self, input: "Input", number: int) -> List["DatumInfo"]:
            """Returns the first batch of datums for the provided input spec.
            Must be called once before `next_batch`.

            Parameters
            ----------
            input : pps.Input
                The input spec.
            number : int
                The number of datums to return. If 0, the default batch size set
                server-side is returned.

            Returns
            -------
            List[pps.DatumInfo]
                List length may be less than `number` if the server has returned
                all datums.
            """
            if self.done:
                return []

            self.queue.put(
                CreateDatumRequest(
                    start=StartCreateDatumRequest(input=input, number=number)
                )
            )

            return self._collect_datums(number)

        def next_batch(self, number: int) -> List["DatumInfo"]:
            """Returns subsequent batches of datums. `first_batch` must have
            been called first.

            Parameters
            ----------
            number : int
                The number of datums to return. If 0, the default batch size set
                server-side is returned.

            Returns
            -------
            List[pps.DatumInfo]
                List length may be less than `number` if the server has returned
                all datums.
            """
            if self.done:
                return []

            self.queue.put(
                CreateDatumRequest(continue_=ContinueCreateDatumRequest(number=number))
            )

            return self._collect_datums(number)

        def _collect_datums(self, number: int) -> List["DatumInfo"]:
            dis = []

            for di in self.response_it:
                dis.append(di)
                if len(dis) == number:
                    break

            # Server has returned all datums
            if len(dis) < number:
                self.done = True

            return dis

        def __next__(self):
            if self.done:
                # Server returned all datums so terminate request iterator
                raise StopIteration

            return self.queue.get()

        def __iter__(self):
            return self

    def create_datum(self) -> "CreateDatumClient":
        """Generates datums for a given input spec.

        Returns
        -------
        CreateDatumClient
        """
        return self.CreateDatumClient(super().create_datum)
