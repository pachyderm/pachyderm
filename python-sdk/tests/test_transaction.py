import grpc

from tests.fixtures import *

from pachyderm_sdk.api import pfs, transaction
from pachyderm_sdk.errors import InvalidTransactionOperation


class TestTransaction:
    """Unit tests for the transaction API."""

    @staticmethod
    def test_batch_transaction(client: TestClient):
        repo_names = [client._generate_name() for _ in range(3)]

        def create_repo_request(repo_name):
            return transaction.TransactionRequest(
                create_repo=pfs.CreateRepoRequest(repo=pfs.Repo.from_uri(repo_name))
            )

        try:
            requests = [create_repo_request(name) for name in repo_names]
            client.transaction.batch_transaction(requests=requests)

            for name in repo_names:
                assert client.pfs.repo_exists(pfs.Repo.from_uri(name))

        finally:
            for name in repo_names:
                client.pfs.delete_repo(repo=pfs.Repo.from_uri(name), force=True)

    @staticmethod
    def test_transaction_context_mgr(client: TestClient):
        with client.transaction.transaction() as txn:
            repo_1 = client.new_repo()
            repo_2 = client.new_repo()

            transactions = client.transaction.list_transaction().transaction_info
            assert len(transactions) == 1
            assert transactions[0].transaction.id == txn.id

            txn_info = client.transaction.inspect_transaction(transaction=txn)
            assert txn_info.transaction.id == txn.id

        assert not client.transaction.transaction_exists(txn)
        assert client.pfs.repo_exists(repo_1)
        assert client.pfs.repo_exists(repo_2)

    @staticmethod
    def test_context_mgr_nested(client: TestClient):
        with client.transaction.transaction() as _txn1:
            assert client.transaction_id is not None
            old_transaction_id = client.transaction_id

            with client.transaction.transaction() as _txn2:
                assert client.transaction_id is not None
                assert client.transaction_id != old_transaction_id

            assert client.transaction_id == old_transaction_id

    @staticmethod
    def test_context_mgr_exception(client: TestClient):
        with pytest.raises(Exception):
            with client.transaction.transaction() as txn:
                repo_1 = client.new_repo()
                repo_2 = client.new_repo()
                raise Exception("oops!")

        assert not client.transaction.transaction_exists(txn)
        assert not client.pfs.repo_exists(repo_1)
        assert not client.pfs.repo_exists(repo_2)

    @staticmethod
    def test_delete_transaction(client: TestClient, default_project: bool):
        txn = client.transaction.start_transaction()
        repo_1 = client.new_repo(default_project)
        repo_2 = client.new_repo(default_project)
        client.transaction.delete_transaction(transaction=txn)

        assert not client.transaction.transaction_exists(txn)
        assert not client.pfs.repo_exists(repo_1)
        assert not client.pfs.repo_exists(repo_2)

        with pytest.raises(grpc.RpcError):
            client.transaction.delete_transaction(transaction=txn)

    @staticmethod
    def test_file_operations_within_transaction(client: TestClient, default_project: bool):
        """Ensure that file operations"""
        repo = client.new_repo(default_project)
        branch = pfs.Branch(repo=repo, name="master")
        test_file = pfs.File.from_uri(f"{branch}:/file.dat")

        with client.transaction.transaction() as _txn:
            repo_txn = client.new_repo(default_project)
            with client.pfs.commit(branch=branch) as commit:
                with pytest.raises(InvalidTransactionOperation):
                    commit.put_file_from_bytes(path=test_file.path, data=b"hello world")
                with pytest.raises(InvalidTransactionOperation):
                    commit.copy_file(src=test_file.path, dst="/new_file.dat")
                with pytest.raises(InvalidTransactionOperation):
                    commit.delete_file(path=test_file.path)

        assert not client.pfs.path_exists(file=test_file), "file should not exist"
        assert client.pfs.repo_exists(repo=repo_txn), "txn resources should exist"
