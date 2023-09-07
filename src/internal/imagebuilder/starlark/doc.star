"""
Module build exposes the build system to Starlark.
"""

def download_file(name: string, by_platform: dict[string, string]):
    """
    Downloads a file for use in future build phases.

    Args:
        name: A name for this download.
        by_platform:  A dict from platform (like "linux/amd64") to a specification of the
          download.
          Each value looks like:
              url: The URL to download.
              digest: The hash that the downloaded file should have; like "blake3:ab1fce4fda78..."

    Returns:
        A ReferenceList.
    """
