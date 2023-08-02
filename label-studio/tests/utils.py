from typing import Optional
from label_studio_sdk import Project


def connect_pachyderm_import_storage(
    project: Project,
    pach_repo: str,
    pach_commit: str,
    *,
    pach_project: Optional[str] = "default",
    pach_branch: Optional[str] = "master",
    pachd_address: Optional[str] = "localhost:80",
    use_blob_urls: Optional[bool] = True,
    title: Optional[str] = "",
    description: Optional[str] = "",
):
    """Connect a Pachyderm repo to Label Studio to use as source storage and import tasks.

    Parameters
    ----------
    project: Project
        Specify the Label Studio project this storage is being added to
    pach_repo: string
        Specify the name of the Pachyderm repo
    pach_commit: string
        Specify the commit ID to use as your data
    pach_project: string
        Optional, 'default' by default. Specify the Pachyderm project your repo exists in
    pach_branch: string
        Optional, 'master' by default. Specify the Pachyderm branch from your repo to use
    pachd_address: string
        Optional, 'localhost:80' by default. Specify the pachd address to connect to
    use_blob_urls: bool
        Optional, true by default. Specify whether your data is raw image or video data, or JSON tasks.
    title: string
        Optional, specify a title for your Pachyderm import storage that appears in Label Studio
    description: string
        Optional, specify a description for your Pachyderm import storage

    Returns
    -------
    dict:
        containing the same fields as in the request and:

    id: int
        Storage ID
    type: str
        Type of storage
    created_at: str
        Creation time
    last_sync: str
        Time last sync finished, can be empty.
    last_sync_count: int
        Number of tasks synced in the last sync

    """

    payload = {
        "pach_repo": pach_repo,
        "pach_commit": pach_commit,
        "pach_project": pach_project,
        "pach_branch": pach_branch,
        "pachd_address": pachd_address,
        "use_blob_urls": use_blob_urls,
        "title": title,
        "description": description,
        "project": project.id,
    }
    response = project.make_request("POST", "/api/storages/pachyderm", json=payload)
    return response.json()
