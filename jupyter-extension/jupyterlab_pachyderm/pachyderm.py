class MountInterface:
    async def list_repos(self):
        pass

    async def list_mounts(self):
        pass

    async def mount(self, body):
        pass

    async def unmount(self, body):
        pass

    async def commit(self, body):
        pass

    async def unmount_all(self):
        pass

    async def mount_datums(self, body):
        pass

    async def show_datum(self, slug):
        pass

    async def get_datums(self):
        pass

    async def config(self, body):
        pass

    async def auth_login(self):
        pass

    async def auth_logout(self):
        pass

    async def health(self):
        pass
