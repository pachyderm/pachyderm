import os

from jupyter_server.services.contents.filemanager import FileContentsManager
from tornado.web import HTTPError


class PFSContentsManager(FileContentsManager):
    def __init__(self, root_dir="/pfs"):
        super().__init__()
        self.root_dir = root_dir

    def _validate_root_dir(self, proposal):
        """Avoid validating the root_dir and potentially crashing the server
        extension because we want to allow the user to create it."""
        pass

    def get(self, path, content=True, type=None, format=None):
        """Wrapper method that overwrites the error message for a 404 error."""
        try:
            return super().get(path, content, type, format)
        except HTTPError as error:
            if error.status_code == 404:
                # Include the full path to the content as the all PFS content is
                #   locally mounted and the mount point is only accessible by the
                #   backend. The frontend is able to then provide a more helpful
                #   error message.
                full_path = os.path.join(self.root_dir, path.lstrip("/"))
                error.log_message = f"file or directory does not exist: {full_path}"
            raise error
