import requests
import pytest

from label_studio_sdk import Client as LSClient
from pachyderm_sdk import Client as PachClient
from pachyderm_sdk.api import pfs

from tests.constants import LABEL_STUDIO_URL, AUTH_TOKEN, PACHD_ADDRESS


## TODO: proper client config for CI
@pytest.fixture(scope='session')
def pach_client():
    return PachClient.from_config()

@pytest.fixture
def pachyderm_data(pach_client):
    import_repo = pfs.Repo(name='images', type='user')
    pach_client.pfs.create_repo(repo=import_repo)

    with pach_client.pfs.commit(
        branch=pfs.Branch(repo=import_repo, name='master')
    ) as commit:
        # TODO: switch out
        # commit.put_file_from_url(
        #     path='/liberty.jpg',
        #     url='https://docs.pachyderm.com/images/opencv/liberty.jpg',
        # )

        f = open('/Users/malyala/Downloads/liberty.jpg', 'rb')
        commit.put_file_from_file(path='/liberty.jpg', file=f)
        commit_id = commit.id

    export_repo = pfs.Repo(name='labels')
    pach_client.pfs.create_repo(repo=export_repo)

    yield import_repo, commit_id, export_repo

    pach_client.pfs.delete_all()

@pytest.fixture
def project():
    ls_client = LSClient(url=LABEL_STUDIO_URL, api_key=AUTH_TOKEN)
    ls_client.check_connection()

    project = ls_client.start_project(
        title='Image Annotation',
        label_config='''
    <View>
        <Image name='image' value='$image'/>
        <RectangleLabels name='label' toName='image'>
            <Label value='Statue' background='green'/>
        </RectangleLabels>
    </View>
    ''',
    )

    return project

@pytest.fixture
def import_storage(pachyderm_data, project):
    import_repo, commit_id, _ = pachyderm_data

    # TODO: pachd address to minikube ip
    # Create import storage
    payload = {
        'pach_repo': import_repo.name,
        'pach_commit': commit_id,
        'pachd_address': PACHD_ADDRESS,
        'use_blob_urls': True,
        'title': 'Pachyderm images',
        'project': project.id,
    }
    storage = project.make_request('POST', '/api/storages/pachyderm', json=payload)

    assert storage.ok
    assert storage.json()['type'] == 'pachyderm'
    storage_id = storage.json()["id"]

    # Get all storages
    response = project.make_request('GET', f'/api/storages/pachyderm?project={storage_id}')

    assert response.ok
    assert len(response.json()) == 1

    yield storage.json()

    # Delete storage
    response = project.make_request('DELETE', f'/api/storages/pachyderm/{storage_id}')

    assert response.ok

    # Get all storages
    response = project.make_request('GET', f'/api/storages/pachyderm?project={storage_id}')

    assert response.ok
    assert len(response.json()) == 0


# Label Studio API: https://labelstud.io/api

def test_pachyderm_import_storage(pachyderm_data, project, import_storage):
    import_repo, commit_id, _ = pachyderm_data

    # Validate storage
    payload = {
        'pach_repo': import_repo.name,
        'pach_commit': commit_id,
        'pachd_address': PACHD_ADDRESS,
        'use_blob_urls': True,
        'title': 'Pachyderm images',
        'project': project.id,
    }
    response = project.make_request('POST', '/api/storages/pachyderm/validate', json=payload)

    assert response.ok

    # Update storage
    payload = {
        'pach_repo': import_repo.name,
        'pach_commit': commit_id,
        'pachd_address': PACHD_ADDRESS,
        'use_blob_urls': True,
        'title': 'Pachyderm images (updated)'
    }
    response = project.make_request('PATCH', f'/api/storages/pachyderm/{import_storage["id"]}?project={project.id}', json=payload)

    assert response.ok
    assert response.json()['title'] == 'Pachyderm images (updated)'

    # Get storage
    response = project.make_request('GET', f'/api/storages/pachyderm/{import_storage["id"]}')

    assert response.ok
    assert response.json()['title'] == 'Pachyderm images (updated)'


def test_begin():
    x = requests.get(
        f'{LABEL_STUDIO_URL}/api/current-user/whoami',
        headers={'Authorization': f'Token {AUTH_TOKEN}'},
    )

    assert x.ok
