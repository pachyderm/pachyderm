import requests
import pytest

from label_studio_sdk import Client as LSClient
from pachyderm_sdk import Client as PachClient
from pachyderm_sdk.api import pfs

from tests.constants import LABEL_STUDIO_URL, AUTH_TOKEN
from tests.utils import connect_pachyderm_import_storage


## TODO: proper client config for CI
@pytest.fixture(scope="session")
def pach_client():
    return PachClient.from_config()


@pytest.fixture
def add_data_to_pachyderm(pach_client):
    import_repo = pfs.Repo(name="images", type="user")
    pach_client.pfs.create_repo(repo=import_repo)

    with pach_client.pfs.commit(
        branch=pfs.Branch(repo=import_repo, name="master")
    ) as commit:
        # TODO: switch out
        # commit.put_file_from_url(
        #     path="/liberty.jpg",
        #     url="https://docs.pachyderm.com/images/opencv/liberty.jpg",
        # )

        f = open("/Users/malyala/Downloads/liberty.jpg", "rb")
        commit.put_file_from_file(path="/liberty.jpg", file=f)

        commit_id = commit.id
        print(commit_id)

    export_repo = pfs.Repo(name="labels")
    pach_client.pfs.create_repo(repo=export_repo)

    yield import_repo, commit_id, export_repo

    pach_client.pfs.delete_all()


def test_import_from_pachyderm_storage(add_data_to_pachyderm):
    import_repo, commit_id, export_repo = add_data_to_pachyderm

    ls_client = LSClient(url=LABEL_STUDIO_URL, api_key=AUTH_TOKEN)
    ls_client.check_connection()

    project = ls_client.start_project(
        title="Image Annotation",
        label_config="""
    <View>
        <Image name="image" value="$image"/>
        <RectangleLabels name="label" toName="image">
            <Label value="Statue" background="green"/>
        </RectangleLabels>
    </View>
    """,
    )

    # TODO: pachd address to minikube ip
    storage = connect_pachyderm_import_storage(
        project,
        import_repo.name,
        commit_id,
        pachd_address="host.docker.internal:30650",
        title="Images",
    )

    print(storage)


def test_begin():
    x = requests.get(
        f"{LABEL_STUDIO_URL}/api/current-user/whoami",
        headers={"Authorization": f"Token {AUTH_TOKEN}"},
    )
    print(x.json())

    assert x.ok
