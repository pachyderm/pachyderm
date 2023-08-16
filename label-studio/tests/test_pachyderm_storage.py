import pytest
import json

from label_studio_sdk import Client as LSClient
from pachyderm_sdk import Client as PachClient
from pachyderm_sdk.api import pfs

from tests.constants import LABEL_STUDIO_URL, AUTH_TOKEN, PACHD_ADDRESS


@pytest.fixture(scope="session")
def pach_client():
    """Returns a Pachyderm client connected to a pachd instance."""
    return PachClient.from_pachd_address(PACHD_ADDRESS)


@pytest.fixture
def pachyderm_data(pach_client):
    """
    Creates two Pachyderm repos, images and labels. Images contains an image
    file to be annotated in Label Studio. Labels is the output repo that
    stores Label Studio annotations. Yields the two repos and the commit ID
    associated with the image file.
    """

    import_repo = pfs.Repo(name="images", type="user")
    pach_client.pfs.create_repo(repo=import_repo)

    with pach_client.pfs.commit(
        branch=pfs.Branch(repo=import_repo, name="master")
    ) as commit:
        commit.put_file_from_url(
            path="/liberty.jpg",
            url="https://docs.pachyderm.com/images/opencv/liberty.jpg",
        )
        commit_id = commit.id

    export_repo = pfs.Repo(name="labels", type="user")
    pach_client.pfs.create_repo(repo=export_repo)
    branch = pfs.Branch(repo=export_repo, name="master")
    pach_client.pfs.create_branch(branch=branch)

    yield import_repo, commit_id, export_repo

    pach_client.pfs.delete_all()


@pytest.fixture(scope="session")
def project():
    """Creates and returns a Label Studio project."""
    ls_client = LSClient(url=LABEL_STUDIO_URL, api_key=AUTH_TOKEN)
    ls_client.check_connection()

    project = ls_client.start_project(
        title="Image Annotation",
        label_config="""
    <View>
        <Image name='image' value='$image'/>
        <RectangleLabels name='label' toName='image'>
            <Label value='Statue' background='green'/>
        </RectangleLabels>
    </View>
    """,
    )

    return project


@pytest.fixture
def import_storage_id(pachyderm_data, project):
    """
    Adds a Pachyderm import storage to the Label Studio project and yields the
    ID associated with the storage.
    """

    import_repo, commit_id, _ = pachyderm_data

    # Create import storage
    payload = {
        "pach_repo": import_repo.name,
        "pach_commit": commit_id,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Images",
        "project": project.id,
    }
    response = project.make_request("POST", "/api/storages/pachyderm", json=payload)
    storage_id = response.json()["id"]
    project_id = response.json()["project"]

    # Get all storages
    response = project.make_request(
        "GET", f"/api/storages/pachyderm?project={project_id}"
    )
    assert len(response.json()) == 1

    yield storage_id

    # Delete storage
    project.make_request("DELETE", f"/api/storages/pachyderm/{storage_id}")

    # Get all storages
    response = project.make_request(
        "GET", f"/api/storages/pachyderm?project={project_id}"
    )
    assert len(response.json()) == 0


@pytest.fixture
def export_storage_id(pachyderm_data, project):
    """
    Adds a Pachyderm export storage to the Label Studio project and yields the
    ID associated with the storage.
    """

    _, _, export_repo = pachyderm_data

    # Create export storage
    payload = {
        "pach_repo": export_repo.name,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Pachyderm labels",
        "project": project.id,
    }
    response = project.make_request(
        "POST", "/api/storages/export/pachyderm", json=payload
    )
    storage_id = response.json()["id"]
    project_id = response.json()["project"]

    # Get all storages
    response = project.make_request(
        "GET", f"/api/storages/export/pachyderm?project={project_id}"
    )
    assert len(response.json()) == 1

    yield storage_id

    # Delete storage
    project.make_request("DELETE", f"/api/storages/export/pachyderm/{storage_id}")

    # Get all storages
    response = project.make_request(
        "GET", f"/api/storages/export/pachyderm?project={project_id}"
    )
    assert len(response.json()) == 0


def test_pachyderm_import_storage(pachyderm_data, project, import_storage_id):
    """
    Tests endpoints associated with a Pachyderm import storage. See API here:
    https://labelstud.io/api.
    """

    import_repo, commit_id, _ = pachyderm_data

    # Validate storage
    payload = {
        "pach_repo": import_repo.name,
        "pach_commit": commit_id,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Images",
        "project": project.id,
    }
    project.make_request("POST", "/api/storages/pachyderm/validate", json=payload)

    # Update storage
    payload = {
        "pach_repo": import_repo.name,
        "pach_commit": commit_id,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Images (updated)",
    }
    response = project.make_request(
        "PATCH", f"/api/storages/pachyderm/{import_storage_id}", json=payload
    )
    assert response.json()["title"] == "Images (updated)"

    # Get storage
    response = project.make_request(
        "GET", f"/api/storages/pachyderm/{import_storage_id}"
    )
    assert response.json()["title"] == "Images (updated)"


def test_pachyderm_export_storage(pachyderm_data, project, export_storage_id):
    """
    Tests endpoints associated with a Pachyderm export storage. See API here:
    https://labelstud.io/api.
    """
    _, _, export_repo = pachyderm_data

    # Validate storage
    payload = {
        "pach_repo": export_repo.name,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Labels",
        "project": project.id,
    }
    project.make_request(
        "POST", "/api/storages/export/pachyderm/validate", json=payload
    )

    # Update storage
    payload = {
        "pach_repo": export_repo.name,
        "pachd_address": PACHD_ADDRESS,
        "use_blob_urls": True,
        "title": "Labels (updated)",
    }
    response = project.make_request(
        "PATCH", f"/api/storages/export/pachyderm/{export_storage_id}", json=payload
    )
    assert response.json()["title"] == "Labels (updated)"

    # Get storage
    response = project.make_request(
        "GET", f"/api/storages/export/pachyderm/{export_storage_id}"
    )
    assert response.json()["title"] == "Labels (updated)"


def test_annotate_and_save(
    pachyderm_data, project, import_storage_id, export_storage_id, pach_client
):
    """
    An end-to-end test that reads data from Pachyderm import storage, annotates
    the data, and saves the annotation to Pachyderm export storage.
    """

    _, _, export_repo = pachyderm_data

    project.make_request("POST", f"/api/storages/pachyderm/{import_storage_id}/sync")

    task_id = project.tasks[0]["id"]
    result = [
        {
            "original_width": 1280,
            "original_height": 853,
            "image_rotation": 0,
            "value": {
                "x": 27.687471665927234,
                "y": 4.645161290322584,
                "width": 36.544858870967744,
                "height": 95.35483870967742,
                "rotation": 0,
                "rectanglelabels": ["Statue"],
            },
            "id": "L-ZO8nCsBf",
            "from_name": "label",
            "to_name": "image",
            "type": "rectanglelabels",
            "origin": "manual",
        }
    ]
    project.create_annotation(task_id, result=result)

    project.make_request(
        "POST", f"/api/storages/export/pachyderm/{export_storage_id}/sync"
    )

    file = pfs.File.from_uri(f"{export_repo.name}@master:/{task_id}")
    with pach_client.pfs.pfs_file(file=file) as pfs_file:
        label = json.load(pfs_file)
        assert label["id"] == task_id
