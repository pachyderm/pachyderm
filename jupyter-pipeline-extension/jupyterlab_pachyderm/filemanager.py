from jupyter_server.services.contents.filemanager import FileContentsManager


class PFSContentsManager(FileContentsManager):
    def __init__(self, root_dir="/pfs"):
        super().__init__()
        self.root_dir = root_dir

    def _validate_root_dir(self, proposal):
        """Avoid validating the root_dir and potentially crashing the server
        extension because we want to allow the user to create it."""
        pass
