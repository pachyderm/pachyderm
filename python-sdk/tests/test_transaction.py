import grpc

from tests.fixtures import *
from tests.utils import count

from pachyderm_sdk.api import pfs, transaction


class TestTransaction:
    """Unit tests for the transaction API."""

    @staticmethod
    def test_batch_transaction(client: TestClient):
        expected_repo_count = count(client.pfs.list_repo()) + 3
        repo_names = [client._generate_name() for _ in range(3)]

        def create_repo_request(repo_name):
            return transaction.TransactionRequest(
                create_repo=pfs.CreateRepoRequest(
                    repo=pfs.Repo.from_uri(repo_name)
                )
            )

        try:
            client.transaction.batch_transaction(
                requests=[
                    create_repo_request(name) for name in repo_names
                ]
            )

            assert len(client.transaction.list_transaction().transaction_info) == 0
            assert count(client.pfs.list_repo()) == expected_repo_count

        finally:
            for name in repo_names:
                client.pfs.delete_repo(repo=pfs.Repo.from_uri(name), force=True)

    @staticmethod
    def test_transaction_context_mgr(client: TestClient):
        expected_repo_count = count(client.pfs.list_repo()) + 2

        with client.transaction.transaction() as txn:
            _repo_1 = client.new_repo()
            _repo_2 = client.new_repo()

            transactions = client.transaction.list_transaction().transaction_info
            assert len(transactions) == 1
            assert transactions[0].transaction.id == txn.id

            txn_info = client.transaction.inspect_transaction(transaction=txn)
            assert txn_info.transaction.id == txn.id

        assert count(client.transaction.list_transaction().transaction_info) == 0
        assert count(client.pfs.list_repo()) == expected_repo_count

    @staticmethod
    def test_transaction_context_mgr_nested(client: TestClient):
        with client.transaction.transaction() as _txn1:
            assert client.transaction_id is not None
            old_transaction_id = client.transaction_id

            with client.transaction.transaction() as _txn2:
                assert client.transaction_id is not None
                assert client.transaction_id != old_transaction_id

            assert client.transaction_id == old_transaction_id

    @staticmethod
    def test_transaction_context_mgr_exception(client: TestClient):
        expected_repo_count = count(client.pfs.list_repo())

        print(client._metadata)
        with pytest.raises(Exception):
            with client.transaction.transaction():
                print(client._metadata)
                client.new_repo()
                client.new_repo()
                raise Exception("oops!")

        assert len(client.transaction.list_transaction().transaction_info) == 0
        assert count(client.pfs.list_repo()) == expected_repo_count

    @staticmethod
    def test_delete_transaction(client: TestClient, default_project: bool):
        expected_repo_count = count(client.pfs.list_repo())

        txn = client.transaction.start_transaction()
        client.new_repo(default_project)
        client.new_repo(default_project)
        client.transaction.delete_transaction(transaction=txn)

        assert len(client.transaction.list_transaction().transaction_info) == 0
        assert count(client.pfs.list_repo()) == expected_repo_count

        with pytest.raises(grpc.RpcError):
            client.transaction.delete_transaction(transaction=txn)
