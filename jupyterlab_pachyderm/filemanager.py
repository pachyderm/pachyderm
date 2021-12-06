from jupyter_server.services.contents.filemanager import FileContentsManager


class PFSContentsManager(FileContentsManager):
    def __init__(self, root_dir="/pfs"):
        super().__init__()
        self.root_dir = root_dir
