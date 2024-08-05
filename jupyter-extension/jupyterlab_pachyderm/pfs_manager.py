import datetime
import grpc
from jupyter_server.services.contents.filemanager import FileContentsManager
import mimetypes
from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs, pps, auth
from pachyderm_sdk.api.pfs import Branch, Repo
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
    file_uri: typing.Optional[str]
    """PFS File URI included when listing the contents of directories for use with pagination"""


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
        file_uri=str(fileinfo.file),
    )


def _get_file_model(
    client: Client, model: ContentModel, path: str, file: pfs.File, format: str
):
    model["mimetype"] = mimetypes.guess_type(path)[0]
    with client.pfs.pfs_file(file=file) as pfs_file:
        value = pfs_file.readall()

    if not format:
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
    client: Client, path: str, file: pfs.File, pagination_marker: pfs.File, number
) -> (typing.List[ContentModel], bool):
    list_response = client.pfs.list_file(
        file=file, pagination_marker=pagination_marker, number=number
    )
    dir_contents = []
    for file_info in list_response:
        name = Path(file_info.file.path).name
        model = ContentModel(
            name=name,
            path=str(Path(path, name)),
            type="directory" if file_info.file_type == pfs.FileType.DIR else "file",
            created=file_info.committed,
            last_modified=file_info.committed,
            content=None,
            mimetype=None,
            format=None,
            writable=False,
            file_uri=str(file_info.file),
        )
        dir_contents.append(model)
    return dir_contents


def _get_dir_model(
    client: Client,
    model: ContentModel,
    path: str,
    file: pfs.File,
    pagination_marker: pfs.File,
    number: int,
):
    model["content"] = _create_dir_content(
        client=client,
        path=path,
        file=file,
        pagination_marker=pagination_marker,
        number=number,
    )
    model["mimetype"] = None
    model["format"] = "json"


def _download_file(client: Client, file: pfs.File, destination: Path):
    """Downloads the given PFS file or directory to the CWD"""
    fileinfo = client.pfs.inspect_file(file=file)
    if fileinfo.file_type == pfs.FileType.FILE:
        local_file = destination.joinpath(os.path.basename(file.path))
        with client.pfs.pfs_file(file=file) as src_file:
            with local_file.open("xb") as dst_file:
                shutil.copyfileobj(fsrc=src_file, fdst=dst_file)
    elif fileinfo.file_type == pfs.FileType.DIR:
        if file.path == "/" or file.path == "":
            dir_name = file.commit.repo.name
        else:
            dir_name = os.path.basename(file.path)

        local_path = destination.joinpath(dir_name)
        if local_path.exists():
            raise FileExistsError(local_path)

        with client.pfs.pfs_tar_file(file=file) as tar:
            tar.extractall(path=str(local_path))
    else:
        raise ValueError(
            f"Downloading {file.path} which is unsupported file type {fileinfo.file_type}"
        )


def _default_name(branch: pfs.Branch) -> str:
    name = branch.repo.name
    if branch.repo.project.name and branch.repo.project.name != "default":
        name = f"{branch.repo.project.name}_{name}"
    if branch.name and branch.name != "master":
        name = f"{name}_{branch.name}"
    return name

def _default_commit_name(commit: pfs.Commit) -> str:
    name = commit.repo.name
    if commit.repo.project.name and commit.repo.project.name != "default":
        name = f"{commit.repo.project.name}_{name}"
    if commit.branch.name and commit.branch.name != "master":
        name = f"{name}_{commit.branch.name}"
    return name

class PFSManager(FileContentsManager):
    mounted_commit: typing.Optional[pfs.Commit] = None

    class Repo(typing.TypedDict):
        name: str
        project: str
        uri: str
        branches: typing.List[Branch]

    class Branch(typing.TypedDict):
        name: str
        uri: str

    def __init__(self, client: Client, **kwargs):
        self._client = client
        super().__init__(**kwargs)

    def mount_commit(self, commit: pfs.Commit):
        if not self.commit_exists(commit=commit):
            raise ValueError("commit_uri exists but does not resolve a valid branch")
        self.mounted_commit = commit

    def unmount_commit(self):
        self.mounted_commit = None

    def list_repos(self) -> typing.List[Repo]:
        return {
            r.repo.as_uri(): self.Repo(
                name=r.repo.name,
                project=r.repo.project.name,
                uri=r.repo.as_uri(),
                branches=[self.Branch(name=b.name, uri=b.as_uri()) for b in r.branches],
            )
            for r in self._client.pfs.list_repo()
        }

    def commit_exists(self, commit: pfs.Commit) -> bool:
        try:
            return self._client.pfs.inspect_commit(commit=commit) is not None
        except:
            return False

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

    # returns None for empty path, i.e. the top-level directory or if no branch is mounted
    def _get_file_from_path(self, path: str) -> typing.Optional[pfs.File]:
        if not self.mounted_commit:
            return None

        name = self._get_name(path)
        if not name:
            return None

        path_str = self._get_path(path)
        file_uri = f"{self.mounted_commit.as_uri()}:/{path_str}"
        return pfs.File.from_uri(file_uri)

    def download_file(self, path: str):
        file = self._get_file_from_path(path=path)
        if file is None:
            raise ValueError("Cannot find file to download")
        _download_file(client=self._client, file=file, destination=Path(self.root_dir))

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

    def _get_content_model(self, name: str, path: str, created: datetime, content: typing.Optional[typing.List[ContentModel]]):
        return ContentModel(
            name=name,
            path=path,
            type="directory",
            created=created,
            last_modified=created,
            content=content,
            mimetype=None,
            format="json" if content is not None else None,
            writable=False,
            file_uri=None,
        )

    def get(
        self,
        path,
        content=True,
        type=None,
        format=None,
        pagination_marker: pfs.File = None,
        number: int = None,
    ) -> ContentModel:
        if type == "notebook":
            raise web.HTTPError(
                400,
                "notebook types are not supported in the pachyderm file extension",
                reason="bad type",
            )

        # Show an empty toplevel model if no branch is specified
        if self.mounted_commit is None:
            return self._get_content_model("", "/", DEFAULT_DATETIME, [] if content else None)

        mounted_branch_name = _default_commit_name(self.mounted_commit)
        mounted_branch_created = self._client.pfs.inspect_commit(
            commit=self.mounted_commit
        ).started
        file = self._get_file_from_path(path)

        # Show a toplevel domain with a directory for the mounted branch
        if file is None:
            mounted_branch_content_model = [self._get_content_model(mounted_branch_name, mounted_branch_name, mounted_branch_created, None)] if content else None
            return self._get_content_model("", "/", DEFAULT_DATETIME, mounted_branch_content_model)
        try:
            fileinfo = self._client.pfs.inspect_file(file=file)
        except grpc.RpcError as err:
            # handle getting empty repo root or repo with no commits
            if (
                err.code() == grpc.StatusCode.NOT_FOUND
                or err.code() == grpc.StatusCode.UNKNOWN
            ) and not self._get_path(path=path):
                repo = self._client.pfs.inspect_repo(repo=self.mounted_commit.repo)
                return self._get_content_model(self._get_name(path), path, repo.created, [mounted_branch_content_model])
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
                _get_dir_model(
                    client=self._client,
                    model=model,
                    path=path,
                    file=file,
                    pagination_marker=pagination_marker,
                    number=number,
                )

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
        num_datums_received: int
        all_datums_received: bool

    class CurrentDatum(typing.TypedDict):
        num_datums_received: int
        input: str
        idx: int
        all_datums_received: bool

    _FILEINFO_DIR = os.path.expanduser("~") + "/.cache/pfs_datum"
    _DOWNLOAD_DIR = os.path.expanduser("~") + "/.cache/pfs_datum_download"
    DATUM_BATCH_SIZE = 100

    def __init__(self, client: Client, **kwargs):
        self._client = client
        self._reset()
        super().__init__(**kwargs)

    def _reset(self):
        self._datum_batch_generator = None
        self._datum_list = list()
        self._dirs = set()
        self._datum_index = 0
        self._all_datums_received = False
        self._mount_time = DEFAULT_DATETIME
        self._input = None
        self._download_dir = None
        self._repo_names = {}
        shutil.rmtree(f"{self._FILEINFO_DIR}", ignore_errors=True)
        os.makedirs(self._FILEINFO_DIR, exist_ok=True)
        try:
            os.makedirs(PFS_MOUNT_DIR, exist_ok=True)
        except Exception as e:
            get_logger().error(
                f"Could not create mount dir {PFS_MOUNT_DIR}. Try setting the PFS_MOUNT_DIR env var."
            )
            raise e

    def _populate_name_table(self, input: pps.Input):
        """
        Given the input spec, populates the bi-directional name to commit_uri
        mapping of the commits mounted by the input.
        """
        if input.pfs:
            repo = pfs.Repo(
                name=input.pfs.repo,
                project=pfs.Project(name=input.pfs.project or "default"),
            )
            branch = pfs.Branch(repo=repo, name=input.pfs.branch or "master")
            try:
                commit_id = (
                    input.pfs.commit
                    or self._client.pfs.inspect_branch(branch=branch).head.id
                )
            except grpc.RpcError as err:
                if err.code() == grpc.StatusCode.NOT_FOUND:
                    raise ValueError(f"Input contains non-existent branch {branch}")
                raise err

            commit_uri = f"{repo.as_uri()}@{commit_id}"
            if commit_uri in self._repo_names:
                raise ValueError(
                    "Loading multiple instances of the same commit is currently not supported in the extension"
                )
            name = input.pfs.name or _default_name(branch=branch)
            if name in self._repo_names:
                raise ValueError(f"Input contains name collision for name {name}")
            self._repo_names[commit_uri] = name
            self._repo_names[name] = commit_uri
        if input.join:
            for i in input.join:
                self._populate_name_table(input=i)
        if input.group:
            for i in input.group:
                self._populate_name_table(input=i)
        if input.cross:
            for i in input.cross:
                self._populate_name_table(input=i)

    def mount_datums(self, input_dict: dict):
        try:
            input = pps.Input().from_dict(input_dict["input"])
            self._input = input
            self._datum_batch_generator = self._client.pps.generate_datums(
                input_spec=input, batch_size=DatumManager.DATUM_BATCH_SIZE
            )
            self._datum_list = list()
            self._get_datum_batch()
            self._datum_index = 0
            self._mount_time = datetime.datetime.now()
            self._repo_names.clear()
            self._populate_name_table(input=input)
            if len(self._datum_list) == 0:
                self._reset()
                raise ValueError("input produced no datums to mount")
            self._update_mount()
        except grpc.RpcError as e:
            if (
                e.code() == grpc.StatusCode.NOT_FOUND
                or e.code() == grpc.StatusCode.UNKNOWN
                and "not found" in e.details()
            ):
                self._reset()
                raise ValueError("non-existent branch or repo in input")
        except Exception as e:
            self._reset()
            raise e

    def _get_datum_batch(self):
        try:
            batch = next(self._datum_batch_generator)
        except StopIteration:
            batch = []

        self._datum_list.extend(batch)
        if len(batch) < self.DATUM_BATCH_SIZE:
            self._all_datums_received = True

    def next_datum(self):
        if (
            self._datum_index == len(self._datum_list) - 1
            and not self._all_datums_received
        ):
            self._get_datum_batch()
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
            num_datums_received=len(self._datum_list),
            all_datums_received=self._all_datums_received,
        )

    def current_datum(self) -> dict:
        return self.CurrentDatum(
            num_datums_received=len(self._datum_list),
            input=self._input.to_json() if self._input else None,
            idx=self._datum_index,
            all_datums_received=self._all_datums_received,
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
            return PFS_MOUNT_DIR
        except Exception as e:
            shutil.rmtree(download_dir, ignore_errors=True)
            raise e

    def _get_relative_path(self, fileinfo: pfs.FileInfo) -> Path:
        commit = fileinfo.file.commit
        commit_uri = f"{commit.repo.as_uri()}@{commit.id}"
        name = self._repo_names[commit_uri]
        return Path(name, fileinfo.file.path.strip("/"))

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
        commit_uri = self._repo_names[parts[0]]

        if len(parts) == 1:
            pach_path = "/"
        else:
            pach_path = str(Path(*parts[1:]))

        file_uri = f"{commit_uri}:{pach_path}"
        return pfs.File.from_uri(file_uri)

    def download_file(self, path: str):
        file = self._get_file_from_path(path=path)
        _download_file(client=self._client, file=file, destination=Path(self.root_dir))

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
        self,
        fileinfo: pfs.FileInfo,
        path: Path,
        content: bool,
        type: str,
        format: str,
        pagination_marker: pfs.FileInfo,
        number: int,
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
                    client=self._client,
                    model=model,
                    path=path,
                    file=fileinfo.file,
                    pagination_marker=pagination_marker,
                    number=number,
                )
        return model

    def _get_parent_dir_model(self, local_path: Path, content: bool) -> ContentModel:
        if content:
            format = "json"
            content_model = [
                self._get_model_from_disk(
                    local_path=p,
                    content=False,
                    type=None,
                    format=None,
                    pagination_marker=None,
                    number=None,
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
            file_uri=None,
        )

    def _get_model_from_disk(
        self,
        local_path: Path,
        content: bool,
        type: str,
        format: str,
        pagination_marker: pfs.File,
        number: int,
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
                pagination_marker=pagination_marker,
                number=number,
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
            file_uri=None,
        )

    def get(
        self,
        path,
        content=True,
        type=None,
        format=None,
        pagination_marker: pfs.File = None,
        number: int = None,
    ) -> ContentModel:
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
                    local_path=fileinfo_path,
                    content=content,
                    type=type,
                    format=format,
                    pagination_marker=pagination_marker,
                    number=number,
                )
        else:
            file = self._get_file_from_path(path=path)
            fileinfo = self._client.pfs.inspect_file(file=file)
            return self._get_model_from_fileinfo(
                fileinfo=fileinfo,
                path=path,
                content=content,
                type=type,
                format=format,
                pagination_marker=pagination_marker,
                number=number,
            )

    def save(self, model, path=""):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def delete_file(self, path):
        raise NotImplementedError("Pachyderm file-browser is read-only")

    def rename_file(self, old_path, new_path):
        raise NotImplementedError("Pachyderm file-browser is read-only")
