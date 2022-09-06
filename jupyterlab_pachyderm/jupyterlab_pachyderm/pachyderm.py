READ_ONLY = "ro"
READ_WRITE = "rw"


class MountInterface:
    async def list(self):
        pass

    async def mount(self, repo, branch, mode, name):
        pass

    async def unmount(self, repo, branch, name):
        pass

    async def unmount_all(self):
        pass

    async def commit(self, repo, branch, name, message):
        pass

    async def config(self, body):
        pass

    async def auth_login(self):
        pass

    async def auth_logout(self):
        pass

    async def health(self):
        pass
