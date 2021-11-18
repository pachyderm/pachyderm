import pytest
import tornado

pytest_plugins = ["jupyter_server.pytest_plugin"]


@pytest.fixture
def jp_server_config():
    return {"ServerApp": {"jpserver_extensions": {"jupyterlab_pachyderm": True}}}


async def test_handler_repos(jp_fetch):
    r = await jp_fetch("/pachyderm/v1/repos")
    assert r.code == 200


async def test_handler_repo(jp_fetch):
    r = await jp_fetch("/pachyderm/v1/repos/images")
    assert r.code == 200


async def test_handler_repo_mount_without_name(jp_fetch):
    with pytest.raises(tornado.httpclient.HTTPClientError):
        await jp_fetch("/pachyderm/v1/repos/images/_mount", method="PUT", body="{}")


async def test_handler_repo_mount(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/_mount",
        method="PUT",
        params={"name": "images"},
        body="{}",
    )
    assert r.code == 200


async def test_handler_repo_mount_with_mode(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/_mount",
        method="PUT",
        params={"name": "images", "mode": "w"},
        body="{}",
    )
    assert r.code == 200


async def test_handler_repo_unmount(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/_unmount",
        method="PUT",
        params={"name": "images"},
        body="{}",
    )
    assert r.code == 200


async def test_handler_repo_branch_mount(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/master/_mount",
        method="PUT",
        params={"name": "images"},
        body="{}",
    )
    assert r.code == 200


async def test_handler_repo_branch_unmount(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/master/_unmount",
        method="PUT",
        params={"name": "images"},
        body="{}",
    )
    assert r.code == 200


async def test_handler_repo_commit(jp_fetch):
    r = await jp_fetch(
        "/pachyderm/v1/repos/images/master/_commit",
        method="POST",
        params={"name": "images"},
        body='{"message": "First commit"}',
    )
    assert r.code == 200
