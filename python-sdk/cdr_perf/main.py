from pathlib import Path
from shutil import rmtree
from time import time
from typing import Iterator, Tuple

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, storage

FILESET_ID = ""
ENCRYPTED_CACHE = Path(__file__).parent / "cache-encrypted"
DECRYPTED_CACHE = Path(__file__).parent / "cache-decrypted"
DESTINATION = Path(__file__).parent / "fileset"


def main():
    client = Client.from_config()
    repo = pfs.Repo(name="cdr_perf")
    branch = pfs.Branch(name="master", repo=repo)

    # Create/Renew Fileset
    if not FILESET_ID:
        fileset_id = client.storage.create_fileset((
            storage.CreateFilesetRequest(
                append_file=storage.AppendFile(path=path, data=data)
            ) for path, data in iter_files(client, branch)
        )).fileset_id
        print(fileset_id)
    else:
        client.storage.renew_fileset(fileset_id=FILESET_ID, ttl_seconds=1800)
        fileset_id = FILESET_ID

    # Fetch the chunks (encrypted & decrypted)
    client.storage.fetch_chunks(
        fileset_id,
        path="/",
        cache_location=ENCRYPTED_CACHE,
        http_host_replacement="localhost:9000",
        encrypted=True,
    )
    client.storage.fetch_chunks(
        fileset_id,
        path="/",
        cache_location=DECRYPTED_CACHE,
        http_host_replacement="localhost:9000",
        encrypted=False,
    )

    # Perform a no cache assembly
    if DESTINATION.exists():
        rmtree(DESTINATION)
    DESTINATION.mkdir()
    print(f"Assembling Fileset (no cache)")
    start = time()
    client.storage.assemble_fileset(
        fileset_id,
        path="/",
        destination=DESTINATION,
        cache_location=None,
        http_host_replacement="localhost:9000",
    )
    no_cache_duration = time() - start
    print("  - duration: {:.2f} seconds".format(no_cache_duration))

    # Perform batch of encrypted cache assemblies
    encrypted_times = []
    for idx in range(20):
        if DESTINATION.exists():
            rmtree(DESTINATION)
        DESTINATION.mkdir()
        print(f"Assembling Encrypted Fileset ({idx + 1:02d})")
        start = time()
        client.storage.assemble_fileset(
            fileset_id,
            path="/",
            destination=DESTINATION,
            cache_location=ENCRYPTED_CACHE,
            fetch_missing_chunks=False,
            http_host_replacement="localhost:9000",
            encrypted=True,
        )
        encrypted_times.append(time() - start)
        print("  - duration: {:.2f} seconds".format(encrypted_times[-1]))

    # Perform batch of decrypted cache assemblies
    decrypted_times = []
    for idx in range(20):
        if DESTINATION.exists():
            rmtree(DESTINATION)
        DESTINATION.mkdir()
        print(f"Assembling Decrypted Fileset ({idx + 1:02d})")
        start = time()
        client.storage.assemble_fileset(
            fileset_id,
            path="/",
            destination=DESTINATION,
            cache_location=DECRYPTED_CACHE,
            fetch_missing_chunks=False,
            http_host_replacement="localhost:9000",
            encrypted=False,
        )
        decrypted_times.append(time() - start)
        print("  - duration: {:.2f} seconds".format(decrypted_times[-1]))

    print(f"No Cache Assembly Time: {no_cache_duration:.2f}")  # ~310 seconds
    print(f"Average Encrypted Assembly Time: {sum(encrypted_times) / len(encrypted_times):.2f}")  # 57.26 (20 run)
    print(f"Average Decrypted Assembly Time: {sum(decrypted_times) / len(decrypted_times):.2f}")  # 17.95 (20 run)


def iter_files(client, branch) -> Iterator[Tuple[str, bytes]]:
    for directory in ("blob", "text"):
        for file in client.pfs.list_file(file=pfs.File.from_uri(f"{branch}:/{directory}")):
            with client.pfs.pfs_file(file.file) as pfs_file:
                chunk = pfs_file.read(18 * 1024 * 1024)
                while len(chunk) > 0:
                    yield file.file.path, chunk
                    chunk = pfs_file.read(18 * 1024 * 1024)


if __name__ == "__main__":
    main()
