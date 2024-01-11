import datetime
import grpc
from jupyter_server.services.contents.filemanager import FileContentsManager
import mimetypes
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps, auth
import os
from pathlib import Path
from tornado import web
import typing
import shutil

from .env import PFS_MOUNT_DIR
from .log import get_logger

DEFAULT_DATETIME = datetime.datetime.utcfromtimestamp(0)


class ContentModel(typing.TypedDict):
    name: str
    """Basename of the entity."""
    path: str
    """Full (API-style) path to the entity."""
    type: str
    """
    The entity type. Either "file" or "directory".
    Storing and retrieving notebooks from Pachyderm is currently not
    supported by the Pachyderm extension.
    """
    created: datetime.datetime
    """Creation date of the entity."""
    last_modified: datetime.datetime
    """Last modified date of the entity."""
    content: typing.Optional[typing.Union[bytes, typing.List["ContentModel"]]]
    """
    For files:
      The content field is always of type unicode. For text-format file
      models, content simply contains the file’s bytes after decoding as
      UTF-8. Non-text (base64) files are read as bytes, base64 encoded,
      and then decoded as UTF-8.
    For directories:
      The content field contains a list of content-free models representing
      the entities in the directory.
    """
    mimetype: typing.Optional[str]
    """The mimetype of the file. Should be None for directories."""
    format: typing.Optional[str]
    """For files, either "text" or "base64". Always "json" for directories."""
    writable: bool
    """
    Whether or not the entity can be written to. Currently, the Pachyderm
    extension is read-only for non-pipeline operations.
    """


def _create_base_model(path: str, fileinfo: pfs.FileInfo, type: str) -> ContentModel:
    if type == "file" and fileinfo.file_type != pfs.FileType.FILE:
        raise web.HTTPError(400, f"{path} is not a file", reason="bad type")
    if type == "directory" and fileinfo.file_type != pfs.FileType.DIR:
        raise web.HTTPError(400, f"{path} is not a directory", reason="bad type")

    return ContentModel(
        name=Path(path).name,
        path=path,
        created=fileinfo.committed,
        last_modified=fileinfo.committed,
        type="directory" if fileinfo.file_type == pfs.FileType.DIR else "file",
        content=None,
        mimetype=None,
        format=None,
        writable=False,
    )


def _get_file_model(
    client: Client, model: ContentModel, path: str, file: pfs.File, format: str
):
    model["mimetype"] = mimetypes.guess_type(path)[0]
    with client.pfs.pfs_file(file=file) as pfs_file:
        value = pfs_file.readall()

    if format == None:
        try:
            decoded = value.decode(encoding="utf-8")
            format = "text"
        except UnicodeError:
            format = "base64"
    elif format == "text":
        decoded = value.decode(encoding="utf-8")

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


def _create_dir_content(
    client: Client, path: str, file: pfs.File
) -> typing.List[ContentModel]:
    list_response = client.pfs.list_file(file=file)
    dir_contents = []
    for i in list_response:
        name = Path(i.file.path).name
        model = ContentModel(
            name=name,
            path=str(Path(path, name)),
            type="directory" if i.file_type == pfs.FileType.DIR else "file",
            created=i.committed,
            last_modified=i.committed,
            content=None,
            mimetype=None,
            format=None,
            writable=False,
        )
        dir_contents.append(model)
    return dir_contents


def _get_dir_model(client: Client, model: ContentModel, path: str, file: pfs.File):
    model["content"] = _create_dir_content(client=client, path=path, file=file)
    model["mimetype"] = None
    model["format"] = "json"


class PFSManager(FileContentsManager):
    # changes from mount-server impl:
    #  - branch name will always be present in the dir name, even if master.
    #    this is to get around the edge case where you have a repo that is
    #    the prefix of another repo plus a branch name

    class Mount(typing.TypedDict):
        name: str
        repo: str
        project: str
        branch: str

    class Repo(typing.TypedDict):
        repo: str
        project: str
        authorization: str
        branches: typing.List[str]

    def __init__(self, client: Client, **kwargs):
        self._client = client
        # maintain a dict/list of repos we have "mounted"
        self._mounted = dict()
        super().__init__(**kwargs)

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
        if name not in self._mounted:
            raise ValueError(f"attempted to unmount {name} which was not mounted.")
        del self._mounted[name]

    def unmount_all(self):
        self._mounted.clear()

    def _get_auth_str(self, auth_info: pfs.AuthInfo) -> str:
        if not auth_info:
            return "off"
        if auth.Permission.REPO_WRITE in auth_info.permissions:
            return "write"
        if (
            auth.Permission.REPO_READ in auth_info.permissions
            and auth.Permission.REPO_LIST_COMMIT in auth_info.permissions
            and auth.Permission.REPO_LIST_BRANCH in auth_info.permissions
            and auth.Permission.REPO_LIST_FILE in auth_info.permissions
        ):
            return "read"
        return "none"

    def list_mounts(self) -> dict:
        mounted: list[self.Mount] = []
        repo_info = {
            r.repo.name: self.Repo(
                repo=r.repo.name,
                project=r.repo.project.name,
                authorization=self._get_auth_str(r.auth_info),
                branches=[b.name for b in r.branches],
            )
            for r in self._client.pfs.list_repo()
        }

        for (name, branch) in self._mounted.items():
            mounted.append(
                self.Mount(
                    name=name,
                    repo=branch.repo.name,
                    project=branch.repo.project.name,
                    branch=branch.name,
                )
            )
            repo_info[branch.repo.name]["branches"].remove(branch.name)
            if len(repo_info[branch.repo.name]["branches"]) == 0:
                del repo_info[branch.repo.name]

        unmounted = [r for r in repo_info.values()]

        return {"mounted": mounted, "unmounted": unmounted}

    def _get_name(self, path: str) -> str:
        path = path.lstrip("/")
        if not path:
            return None
        return Path(path).parts[0]

    def _get_path(self, path: str) -> str:
        path = path.lstrip("/")
        if not path or len(Path(path).parts) == 1:
            return ""
        return str(Path(*Path(path).parts[1:]))

    # returns None for empty path, i.e. the top-level directory
    def _get_file_from_path(self, path: str) -> pfs.File:
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

    def _get_repo_models(self) -> typing.List[ContentModel]:
        models = []
        for repo, branch in self._mounted.items():
            time = self._client.pfs.inspect_commit(
                commit=pfs.Commit(branch=branch, repo=branch.repo)
            ).started
            models.append(
                ContentModel(
                    name=repo,
                    path=repo,
                    type="directory",
                    created=time,
                    last_modified=time,
                    content=None,
                    mimetype=None,
                    format=None,
                    writable=False,
                )
            )
        return models

    def _get_toplevel_model(self, content: bool) -> ContentModel:
        if content:
            format = "json"
            content_model = self._get_repo_models()
        else:
            format = None
            content_model = None
        return ContentModel(
            name="",
            path="/",
            type="directory",
            created=DEFAULT_DATETIME,
            last_modified=DEFAULT_DATETIME,
            content=content_model,
            mimetype=None,
            format=format,
            writable=False,
        )

    def _get_empty_repo_model(self, name: str, content: bool):
        repo = self._client.pfs.inspect_repo(repo=self._mounted[name].repo)
        return ContentModel(
            name=name,
            path=name,
            type="directory",
            created=repo.created,
            last_modified=repo.created,
            content=[] if content else None,
            mimetype=None,
            format="json" if content else None,
            writable=False,
        )

    def get(self, path, content=True, type=None, format=None) -> ContentModel:
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

        try:
            fileinfo = self._client.pfs.inspect_file(file=file)
        except grpc.RpcError as err:
            # handle getting empty repo root or repo with no commits
            if (
                err.code() == grpc.StatusCode.NOT_FOUND
                or err.code() == grpc.StatusCode.UNKNOWN
            ) and not self._get_path(path=path):
                return self._get_empty_repo_model(
                    name=self._get_name(path), content=content
                )
            else:
                raise err

        model = _create_base_model(path=path, fileinfo=fileinfo, type=type)

        if content:
            if model["type"] == "file":
                _get_file_model(
                    client=self._client,
                    model=model,
                    path=path,
                    file=file,
                    format=format,
                )
            else:
                _get_dir_model(client=self._client, model=model, path=path, file=file)

        return model

    def save(self, model, path=""):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def delete_file(self, path):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def rename_file(self, old_path, new_path):
        raise NotImplementedError("Pachyderm file-browser is read-only")


class DatumManager(FileContentsManager):
    """
    The DatumManager keeps FileInfo metadata on disk at _FILEINFO_DIR in order to
    manage what should and should not be displayed in the file browser for any
    given datum. How it works:
    - When a new datum is mounted (this can either be when mount_datums is called
        or when we are cycling through datums for an input spec), the file metadata
        contained in the FileInfo objects within the datum are written to disk.
    - The file structure on disk mimics what should be displayed in the browser.
    - For example, if we have a file in repo "test" at "master" branch within the
        "default" project located at the path "/foo/bar/file.txt", the FileInfo for
        that file would exist serialized on disk at:
        /pfs_datum/default_test_master/foo/bar/file.txt
    - Both files and directories that are mounted are stored this way. What should
        be displayed, however, varies depending on if a file or directory is mounted.
    - For files, both the file itself as well as any parent directories of the file
        should exist in /pfs_datum
    - For directories, any files or directories within that directory should be
        mounted as well.
    """

    class DatumState(typing.TypedDict):
        id: str
        idx: int
        num_datums: int
        all_datums_received: bool

    class CurrentDatum(typing.TypedDict):
        num_datums: int
        input: str
        idx: int
        all_datums_received: bool

    _FILEINFO_DIR = os.path.expanduser("~") + "/.cache/pfs_datum"
    _DOWNLOAD_DIR = os.path.expanduser("~") + "/.cache/pfs_datum_download"
    # currently unsupported stuff (unclear if needed or not):
    #  - crossing repo with itself
    #  - renaming repo level directories
    def __init__(self, client: Client, **kwargs):
        self._client = client
        self._datum_list = list()
        self._dirs = set()
        self._datum_index = 0
        self._mount_time = DEFAULT_DATETIME
        self._input = None
        self._download_dir = None
        shutil.rmtree(f"{self._FILEINFO_DIR}", ignore_errors=True)
        os.makedirs(self._FILEINFO_DIR, exist_ok=True)
        try:
            os.makedirs(PFS_MOUNT_DIR, exist_ok=True)
        except Exception as e:
            get_logger().error(
                f"Could not create mount dir {PFS_MOUNT_DIR}. Try setting the PFS_MOUNT_DIR env var."
            )
            raise e
        super().__init__(**kwargs)

    # TODO: don't ignore name in the input spec
    # right now, when a repo is mounted as part of mounting datum(s), the
    # toplevel directory for the repo is project_repo_branch. we should
    # update this to respect the name in the pfsinput objects being supplied
    # to the mount_datums api. however, this is not that simple as the name
    # can't simply be looked up in the datuminfo. we will need to have some
    # mapping to track which files within a datum belong to which name.
    def mount_datums(self, input_dict: dict):
        input = pps.Input().from_dict(input_dict["input"])
        self._input = input
        self._datum_list = list(self._client.pps.list_datum(input=input))
        self._datum_index = 0
        self._mount_time = datetime.datetime.now()
        if len(self._datum_list) == 0:
            raise ValueError("input produced no datums to mount")
        self._update_mount()

    def next_datum(self):
        self._datum_index = (self._datum_index + 1) % len(self._datum_list)
        self._update_mount()

    def prev_datum(self):
        if self._datum_index == 0:
            self._datum_index = len(self._datum_list)
        self._datum_index = self._datum_index - 1
        self._update_mount()

    def datum_state(self) -> dict:
        return self.DatumState(
            id=self._datum_list[self._datum_index].datum.id,
            idx=self._datum_index,
            num_datums=len(self._datum_list),
            all_datums_received=True,
        )

    def current_datum(self) -> dict:
        return self.CurrentDatum(
            num_datums=len(self._datum_list),
            input=self._input.to_json() if self._input else None,
            idx=self._datum_index,
            all_datums_received=True,
        )

    def download(self):
        """
        Downloads the currently mounted datum to a local cache directory,
        then links the downloaded files to PFS_MOUNT_DIR (/pfs by default).
        Subsequent calls to this method will remove the previously
        downloaded datum and link the newly downloaded ones.
        """
        if (
            len(self._datum_list) == 0
            or len(self._datum_list[self._datum_index].data) == 0
        ):
            raise ValueError("Attempting to download empty or unmounted datum")

        download_dir = (
            f"{self._DOWNLOAD_DIR}/datum-{datetime.datetime.now().isoformat()}"
        )
        try:
            os.makedirs(download_dir, exist_ok=True)
            for fileinfo in self._datum_list[self._datum_index].data:
                path = self._get_download_path(
                    download_dir=download_dir, fileinfo=fileinfo
                )
                os.makedirs(path.parent, exist_ok=True)
                if fileinfo.file_type == pfs.FileType.FILE:
                    # download individual file
                    with self._client.pfs.pfs_file(file=fileinfo.file) as datum_file:
                        with open(path, "wb") as download_file:
                            shutil.copyfileobj(fsrc=datum_file, fdst=download_file)
                elif fileinfo.file_type == pfs.FileType.DIR:
                    # download tarball
                    with self._client.pfs.pfs_tar_file(file=fileinfo.file) as tar:
                        tar.extractall(
                            path=path if fileinfo.file.path == "/" else path.parent
                        )
                else:
                    raise TypeError(
                        f"Attempting to download invalid file type {fileinfo.file_type}"
                    )

            for dir in os.listdir(PFS_MOUNT_DIR):
                os.unlink(Path(PFS_MOUNT_DIR, dir))

            for dir in os.listdir(download_dir):
                os.symlink(
                    src=Path(download_dir, dir),
                    dst=Path(PFS_MOUNT_DIR, dir),
                    target_is_directory=True,
                )

            shutil.rmtree(self._download_dir, ignore_errors=True)
            self._download_dir = download_dir
        except Exception as e:
            shutil.rmtree(download_dir, ignore_errors=True)
            raise e

    def _get_relative_path(self, fileinfo: pfs.FileInfo) -> Path:
        project = fileinfo.file.commit.repo.project.name
        branch = fileinfo.file.commit.branch.name
        repo = fileinfo.file.commit.repo.name
        toplevel = f"{project}_{repo}_{branch}"
        return Path(toplevel, fileinfo.file.path.strip("/"))

    def _get_download_path(self, download_dir: Path, fileinfo: pfs.FileInfo) -> Path:
        return Path(download_dir, self._get_relative_path(fileinfo=fileinfo))

    def _get_fileinfo_path(self, fileinfo: pfs.FileInfo) -> Path:
        return Path(self._FILEINFO_DIR, self._get_relative_path(fileinfo=fileinfo))

    def _update_mount(self):
        shutil.rmtree(f"{self._FILEINFO_DIR}", ignore_errors=True)
        os.makedirs(self._FILEINFO_DIR)
        self._dirs.clear()
        for fileinfo in self._datum_list[self._datum_index].data:
            path = self._get_fileinfo_path(fileinfo=fileinfo)

            if path.exists():
                if path.is_dir():
                    # the dir being mounted is the parent of a file that is already mounted
                    shutil.rmtree(str(path))
                else:
                    return

            os.makedirs(str(path.parent), exist_ok=True)
            with open(path, "w") as f:
                f.write(fileinfo.to_json())
            if fileinfo.file_type == pfs.FileType.DIR:
                self._dirs.add(str(path))

    def _get_file_from_path(self, path: str) -> pfs.File:
        path = path.strip("/")
        if not path:
            return None

        parts = Path(path).parts
        project, repo, branch = parts[0].split("_")
        if len(parts) == 1:
            pach_path = "/"
        else:
            pach_path = str(Path(*parts[1:]))
        file_uri = f"{project}/{repo}@{branch}:{pach_path}"
        return pfs.File.from_uri(file_uri)

    def is_hidden(self, path):
        return False

    def file_exists(self, path) -> bool:
        # two ways a file exists:
        #  - if the file itself is mounted
        #  - if the file is part of a directory that's mounted
        fileinfo_path = Path(self._FILEINFO_DIR, path)

        if fileinfo_path.is_file():
            with open(fileinfo_path, "r") as f:
                fileinfo = pfs.FileInfo().from_json(value=f.read())
            return fileinfo.file_type == pfs.FileType.FILE

        for dir in fileinfo_path.parents:
            if str(dir) in self._dirs:
                file = self._get_file_from_path(path=path)
                try:
                    fileinfo = self._client.pfs.inspect_file(file=file)
                except grpc.RpcError as err:
                    if err.code() == grpc.StatusCode.NOT_FOUND:
                        return False
                    raise err
                return fileinfo.file_type == pfs.FileType.FILE

        return False

    def dir_exists(self, path) -> bool:
        # three ways a dir exists:
        #  - if the dir itself is mounted
        #  - if the dir is part of a directory that's mounted
        #  - if the dir is a parent of a file or directory that's mounted (i.e. if the path exists)
        fileinfo_path = Path(self._FILEINFO_DIR, path)

        if fileinfo_path.exists():
            if fileinfo_path.is_file():
                with open(fileinfo_path, "r") as f:
                    fileinfo = pfs.FileInfo().from_json(value=f.read())
                return fileinfo.file_type == pfs.FileType.DIR
            return True

        for dir in fileinfo_path.parents:
            if str(dir) in self._dirs:
                file = self._get_file_from_path(path=path)
                try:
                    fileinfo = self._client.pfs.inspect_file(file=file)
                except grpc.RpcError as err:
                    if err.code() == grpc.StatusCode.NOT_FOUND:
                        return False
                    raise err
                return fileinfo.file_type == pfs.FileType.DIR

        return False

    def exists(self, path):
        fileinfo_path = Path(self._FILEINFO_DIR, path)

        if fileinfo_path.exists():
            return True

        for dir in fileinfo_path.parents:
            if str(dir) in self._dirs:
                file = self._get_file_from_path(path=path)
                try:
                    fileinfo = self._client.pfs.inspect_file(file=file)
                except grpc.RpcError as err:
                    if err.code() == grpc.StatusCode.NOT_FOUND:
                        return False
                    raise err
                return True

        return False

    # translate a Path in local storage to a path in the contents FS
    # for example, {self._FILEINFO_DIR}/some_repo/foo/bar becomes
    # /some_repo/foo/bar
    def _translate_path(self, path: Path) -> str:
        pathstr = str(path)
        assert pathstr.startswith(self._FILEINFO_DIR)
        return pathstr[len(self._FILEINFO_DIR) :]

    def _get_model_from_fileinfo(
        self, fileinfo: pfs.FileInfo, path: Path, content: bool, type: str, format: str
    ) -> ContentModel:
        model = _create_base_model(path=path, fileinfo=fileinfo, type=type)
        if content:
            if model["type"] == "file":
                _get_file_model(
                    client=self._client,
                    model=model,
                    path=path,
                    file=fileinfo.file,
                    format=format,
                )
            else:
                _get_dir_model(
                    client=self._client, model=model, path=path, file=fileinfo.file
                )
        return model

    def _get_parent_dir_model(self, local_path: Path, content: bool) -> ContentModel:
        if content:
            format = "json"
            content_model = [
                self._get_model_from_disk(
                    local_path=p, content=False, type=None, format=None
                )
                for p in local_path.iterdir()
            ]
        else:
            format = None
            content_model = None
        return ContentModel(
            name=local_path.name,
            path=self._translate_path(path=local_path),
            created=self._mount_time,
            last_modified=self._mount_time,
            content=content_model,
            type="directory",
            mimetype=None,
            format=format,
            writable=False,
        )

    def _get_model_from_disk(
        self, local_path: Path, content: bool, type: str, format: str
    ):
        if local_path.is_dir():
            if type and type != "directory":
                raise web.HTTPError(
                    400, f"{local_path} is a directory, not a {type}", reason="bad type"
                )
            return self._get_parent_dir_model(local_path=local_path, content=content)
        else:
            with open(local_path, "r") as f:
                fileinfo = pfs.FileInfo().from_json(value=f.read())
            return self._get_model_from_fileinfo(
                fileinfo=fileinfo,
                path=self._translate_path(local_path),
                content=content,
                type=type,
                format=format,
            )

    def _get_root_model(self, content: bool) -> ContentModel:
        if content:
            format = "json"
            content_model = [
                self._get_parent_dir_model(local_path=p, content=False)
                for p in Path(self._FILEINFO_DIR).iterdir()
            ]
        else:
            format = None
            content_model = None
        return ContentModel(
            name="",
            path="/",
            type="directory",
            created=DEFAULT_DATETIME,
            last_modified=DEFAULT_DATETIME,
            content=content_model,
            mimetype=None,
            format=format,
            writable=False,
        )

    def get(self, path, content=True, type=None, format=None) -> ContentModel:
        if type == "notebook":
            raise web.HTTPError(
                400,
                "notebook types are not supported in the pachyderm file extension",
                reason="bad type",
            )

        path = path.strip("/")
        if not path:
            return self._get_root_model(content=content)

        fileinfo_path = Path(self._FILEINFO_DIR, path)

        if fileinfo_path.exists():
            if fileinfo_path.is_dir():
                return self._get_parent_dir_model(
                    local_path=fileinfo_path, content=content
                )
            else:
                return self._get_model_from_disk(
                    local_path=fileinfo_path, content=content, type=type, format=format
                )
        else:
            file = self._get_file_from_path(path=path)
            fileinfo = self._client.pfs.inspect_file(file=file)
            return self._get_model_from_fileinfo(
                fileinfo=fileinfo, path=path, content=content, type=type, format=format
            )

    def save(self, model, path=""):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def delete_file(self, path):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def rename_file(self, old_path, new_path):
        raise NotImplementedError("Pachyderm file-browser is read-only")
