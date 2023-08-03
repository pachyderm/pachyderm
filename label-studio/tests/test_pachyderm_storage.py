import requests
import pytest

from label_studio_sdk import Client as LSClient
from pachyderm_sdk import Client as PachClient
from pachyderm_sdk.api import pfs

from tests.constants import LABEL_STUDIO_URL, AUTH_TOKEN, PACHD_ADDRESS


## TODO: proper client config for CI
@pytest.fixture(scope="session")
def pach_client():
    return PachClient.from_config()


@pytest.fixture
def pachyderm_data(pach_client):
    import_repo = pfs.Repo(name="images", type="user")
    pach_client.pfs.create_repo(repo=import_repo)

    with pach_client.pfs.commit(
        branch=pfs.Branch(repo=import_repo, name="master")
    ) as commit:
        # TODO: switch out
        # commit.put_file_from_url(
        #     path='/liberty.jpg',
        #     url='https://docs.pachyderm.com/images/opencv/liberty.jpg',
        # )

        f = open("/Users/malyala/Downloads/liberty.jpg", "rb")
        commit.put_file_from_file(path="/liberty.jpg", file=f)
        commit_id = commit.id

    export_repo = pfs.Repo(name="labels", type="user")
    pach_client.pfs.create_repo(repo=export_repo)
    branch = pfs.Branch(repo=export_repo, name="master")
    pach_client.pfs.create_branch(branch=branch)

    yield import_repo, commit_id, export_repo

    # pach_client.pfs.delete_all()


@pytest.fixture(scope="session")
def project():
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
def import_storage(pachyderm_data, project):
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

    # # Delete storage
    # project.make_request('DELETE', f'/api/storages/pachyderm/{storage_id}')

    # # Get all storages
    # response = project.make_request('GET', f'/api/storages/pachyderm?project={project_id}')
    # assert len(response.json()) == 0


@pytest.fixture
def export_storage(pachyderm_data, project):
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

    # # Delete storage
    # project.make_request("DELETE", f"/api/storages/export/pachyderm/{storage_id}")

    # # Get all storages
    # response = project.make_request(
    #     "GET", f"/api/storages/export/pachyderm?project={project_id}"
    # )
    # assert len(response.json()) == 0


# # Label Studio API: https://labelstud.io/api
# def test_pachyderm_import_storage(pachyderm_data, project, import_storage):
#     import_repo, commit_id, _ = pachyderm_data

#     # Validate storage
#     payload = {
#         'pach_repo': import_repo.name,
#         'pach_commit': commit_id,
#         'pachd_address': PACHD_ADDRESS,
#         'use_blob_urls': True,
#         'title': 'Images',
#         'project': project.id,
#     }
#     project.make_request('POST', '/api/storages/pachyderm/validate', json=payload)

#     # Update storage
#     payload = {
#         'pach_repo': import_repo.name,
#         'pach_commit': commit_id,
#         'pachd_address': PACHD_ADDRESS,
#         'use_blob_urls': True,
#         'title': 'Images (updated)'
#     }
#     response = project.make_request('PATCH', f'/api/storages/pachyderm/{import_storage}', json=payload)
#     assert response.json()['title'] == 'Images (updated)'

#     # Get storage
#     response = project.make_request('GET', f'/api/storages/pachyderm/{import_storage}')
#     assert response.json()['title'] == 'Images (updated)'


# def test_pachyderm_export_storage(pachyderm_data, project, export_storage):
#     _, _, export_repo = pachyderm_data

#     # Validate storage
#     payload = {
#         'pach_repo': export_repo.name,
#         'pachd_address': PACHD_ADDRESS,
#         'use_blob_urls': True,
#         'title': 'Labels',
#         'project': project.id,
#     }
#     project.make_request('POST', '/api/storages/export/pachyderm/validate', json=payload)

#     # Update storage
#     payload = {
#         'pach_repo': export_repo.name,
#         'pachd_address': PACHD_ADDRESS,
#         'use_blob_urls': True,
#         'title': 'Labels (updated)'
#     }
#     response = project.make_request('PATCH', f'/api/storages/export/pachyderm/{export_storage}', json=payload)
#     assert response.json()['title'] == 'Labels (updated)'

#     # Get storage
#     response = project.make_request('GET', f'/api/storages/export/pachyderm/{export_storage}')
#     assert response.json()['title'] == 'Labels (updated)'


def test_annotate_and_save(project, import_storage, export_storage, pach_client):
    project.make_request("POST", f"/api/storages/pachyderm/{import_storage}/sync")

    task_id = project.tasks[0]["id"]
    result = [{
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
    }]
    project.create_annotation(task_id, result=result)

    project.make_request("POST", f"/api/storages/export/pachyderm/{export_storage}/sync")
    
    # pach_client.