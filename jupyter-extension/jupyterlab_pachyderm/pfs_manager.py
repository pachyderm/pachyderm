import datetime
import grpc
from jupyter_server.services.contents.filemanager import FileContentsManager
import mimetypes
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs
from pathlib import Path
from tornado import web


class PFSManager(FileContentsManager):
    # changes from mount-server impl:
    #  - branch name will always be present in the dir name, even if master.
    #    this is to get around the edge case where you have a repo that is
    #    the prefix of another repo plus a branch name

    # maintain a dict/list of repos we have "mounted"
    def __init__(self, client: Client, **kwargs):
        self._client = client
        self._mounted = dict()
        # until we add a way to mount in the UI, you will need to add code to mount
        # self.mount_repo(repo="test", branch="master")
        # self.mount_repo(repo="test_testing", branch="master")

    def mount_repo(
        self, repo: str, branch: str, project: str = "default", name: str = None
    ):  # maybe name should be required?
        # let's assume for now that there are no '_' in any names
        # TODO: we need to figure out a scheme to get around naming
        # conflicts in the top level directory
        if not name:
            name = f"{project}_{repo}_{branch}"
        if name in self._mounted:
            raise ValueError(f"attempted to mount as {name} which is already mounted.")
        branch_uri = f"{project}/{repo}@{branch}"
        mounted_branch = pfs.Branch.from_uri(branch_uri)
        self._mounted[name] = mounted_branch

    def unmount_repo(self, name: str):
        if not self._mounted[name]:
            raise ValueError(f"attempted to unmount {name} which was not mounted.")
        del self._mounted[name]

    def _get_name(self, path: Path) -> str:
        if not path:
            return None
        return Path(path).parts[0]

    def _get_path(self, path: Path) -> str:
        if not path or len(Path(path).parts) == 1:
            return ""
        return str(Path(*Path(path).parts[1:]))

    # returns None for empty path, i.e. the top-level directory
    def _get_file_from_path(self, path: str) -> pfs.File:
        path = path.lstrip("/")
        name = self._get_name(path)
        if not name:
            return None

        if name not in self._mounted:
            raise FileNotFoundError(f"{name} not mounted")

        path_str = self._get_path(path)
        file_uri = f"{self._mounted[name]}:/{path_str}"
        return pfs.File.from_uri(file_uri)

    def is_hidden(self, path):
        return False

    def file_exists(self, path) -> bool:
        file = self._get_file_from_path(path)
        if file is None:
            return False  # top-level is always a directory
        try:
            response = self._client.pfs.inspect_file(file=file)
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err
        return response.file_type == pfs.FileType.FILE

    def dir_exists(self, path):
        file = self._get_file_from_path(path)
        if file is None:
            return True  # top-level is always a directory
        try:
            response = self._client.pfs.inspect_file(file=file)
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.NOT_FOUND:
                return False
            raise err
        return response.file_type == pfs.FileType.DIR

    def exists(self, path) -> bool:
        file = self._get_file_from_path(path)
        if file is None:
            return True
        return self._client.pfs.path_exists(file=file)

    def _create_base_model(self, path: str, fileinfo: pfs.FileInfo, type: str) -> dict:
        if type == "file" and fileinfo.file_type != pfs.FileType.FILE:
            raise web.HTTPError(400, f"{path} is not a file", reason="bad type")
        if type == "directory" and fileinfo.file_type != pfs.FileType.DIR:
            raise web.HTTPError(400, f"{path} is not a directory", reason="bad type")

        model = {}
        model["name"] = Path(path).name
        model["path"] = path
        model["created"] = fileinfo.committed
        model["last_modified"] = fileinfo.committed
        model["type"] = (
            "directory" if fileinfo.file_type == pfs.FileType.DIR else "file"
        )
        model["content"] = None
        model["mimetype"] = None
        model["format"] = None
        model["writable"] = False

        return model

    def _get_file_model(self, model: dict, path: str, file: pfs.File, format: str):
        get_response = self._client.pfs.get_file(file=file)
        model["mimetype"] = mimetypes.guess_type(path)[0]
        for content in get_response:
            value = content.value

        if format == None:
            try:
                decoded = value.decode(encoding="utf-8")
                format = "text"
            except UnicodeError:
                format = "base64"

        if format == "text":
            model["format"] = "text"
            model["content"] = decoded
            if model["mimetype"] is None:
                model["mimetype"] = "text/plain"
        else:
            model["format"] = "base64"
            model["content"] = value
            if model["mimetype"] is None:
                model["mimetype"] = "application/octet-stream"

    def _create_dir_content(self, path: str, file: pfs.File) -> list:
        list_response = self._client.pfs.list_file(file=file)
        dir_contents = []
        for i in list_response:
            name = Path(i.file.path).name
            model = {}
            model["name"] = name
            model["path"] = str(Path(path, name))
            model["type"] = "directory" if i.file_type == pfs.FileType.DIR else "file"
            model["created"] = i.committed
            model["last_modified"] = i.committed
            model["content"] = None
            model["mimetype"] = None
            model["format"] = None
            model["writable"] = False
            dir_contents.append(model)
        return dir_contents

    def _get_dir_model(self, model: dict, path: str, file: pfs.File):
        model["content"] = self._create_dir_content(path=path, file=file)
        model["mimetype"] = None
        model["format"] = "json"

    def _get_repo_model(self, repo: str) -> dict:
        branch = self._mounted[repo]
        head = self._client.pfs.inspect_branch(branch=branch).head
        time = self._client.pfs.inspect_commit(commit=head).finished
        model = {}
        model["name"] = repo
        model["path"] = repo
        model["type"] = "directory"
        model["created"] = time
        model["last_modified"] = time
        model["content"] = None
        model["mimetype"] = None
        model["format"] = None
        model["writable"] = False
        return model

    def _get_toplevel_model(self, content: bool) -> dict:
        model = {}
        model["name"] = ""
        model["path"] = "/"
        model["type"] = "directory"
        model["created"] = datetime.datetime.min
        model["last_modified"] = datetime.datetime.min
        model["content"] = None
        model["mimetype"] = None
        model["format"] = None
        model["writable"] = False
        if content:
            model["format"] = "json"
            model["content"] = [self._get_repo_model(r) for r in self._mounted]
        return model

    def get(self, path, content=True, type=None, format=None) -> dict:
        if type == "notebook":
            raise web.HTTPError(
                400,
                "notebook types are not supported in the pachyderm file extension",
                reason="bad type",
            )

        file = self._get_file_from_path(path=path)

        if file is None:
            # show top-level dir
            return self._get_toplevel_model(content=content)

        fileinfo = self._client.pfs.inspect_file(file=file)

        model = self._create_base_model(path=path, fileinfo=fileinfo, type=type)

        if content:
            if model["type"] == "file":
                self._get_file_model(model=model, path=path, file=file, format=format)
            else:
                self._get_dir_model(model=model, path=path, file=file)

        return model

    def save(self, model, path=""):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def delete_file(self, path):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def rename_file(self, old_path, new_path):
        raise NotImplementedError("Pachyderm file-browser is read-only")
