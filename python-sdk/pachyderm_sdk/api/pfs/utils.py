from shutil import which
from subprocess import run


def check_pachctl(*, ensure_mount: bool = False) -> None:
    """Ensures that pachctl is installed and is capable of running
    mount operations.

    Errors
    ------
    FileNotFoundError
        No pachctl binary present
    RuntimeError
        pachctl binary present but cannot run mount operations
    """
    if which("pachctl") is None:
        raise FileNotFoundError("pachctl")
    if ensure_mount:
        if run(["pachctl", "mount", "-h"], capture_output=True).returncode != 0:
            raise RuntimeError("incompatible pachctl detected")
