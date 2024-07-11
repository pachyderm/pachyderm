import faker
import random
from tqdm import tqdm

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs

SEED = int(b"PACHYDERM".hex(), 16)
GB = 1024 ** 3
DATASET_SIZE = 10 * GB
BAR_FORMAT = "{l_bar}{bar}| {n:0.2f}/{total_fmt} [{elapsed}<{remaining}, ' '{rate_fmt}{postfix}]"


def main():
    client = Client.from_config()
    repo = pfs.Repo(name="cdr_perf")
    branch = pfs.Branch(name="master", repo=repo)
    if client.pfs.repo_exists(repo):
        print("Repo already exists")
        exit(1)
    client.pfs.create_repo(repo=repo)

    generate_text_dataset(client, branch)
    generate_blob_dataset(client, branch)


def generate_text_dataset(client, branch):
    faker.Faker.seed(SEED)
    fake = faker.Faker()

    count = 0
    size = 0
    with tqdm(total=DATASET_SIZE / GB, desc="Generating text files", bar_format=BAR_FORMAT) as bar:
        with client.pfs.commit(branch=branch) as commit:
            while size < DATASET_SIZE:
                text = fake.text(random.randint(100, 10_000_000)).encode("utf-8")
                commit.put_file_from_bytes(f"/text/{count:06d}", text)
                size += len(text)
                count += 1
                bar.update(len(text) / GB)


def generate_blob_dataset(client, branch):
    random.seed(SEED)

    count = 0
    size = 0
    with tqdm(total=DATASET_SIZE / GB, desc="Generating blob files", bar_format=BAR_FORMAT) as bar:
        with client.pfs.commit(branch=branch) as commit:
            while size < DATASET_SIZE:
                blob = random.randbytes(random.randint(1000, 100_000_000))
                commit.put_file_from_bytes(f"/blob/{count:06d}", blob)
                size += len(blob)
                count += 1
                bar.update(len(blob) / GB)


if __name__ == "__main__":
    main()
