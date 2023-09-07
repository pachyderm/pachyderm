"""
Module build exposes the build system to Starlark.  All methods configure work to be done;
they do not acutally do the work immediately.
"""

# Do not import.
load("doc.star", "*")

def ReferenceList():
    # Actually a type, not a function.
    """
    A ReferenceList is a set of references.  References point at build targets, and can
    be passed to the build system to yield the desired artifact.

    A ReferenceList can be used as a dict, list, or iterator.

    When indexing as a dict, the key is used as a selector for references.
    "#linux/amd64" matches all references that select a platform of linux/amd64.  "foo" selects
    references that matches a target named "foo".  "foo#linux/amd64" matches references that
    match targets with the name "foo" AND the platform "linux/amd64".
    """

def path(selector):
    """
    Creates a dependency on a filesystem path, relative to the script file that is executing this
    statement.

    Args:
        selector: A string pointing at a file or a directory.  Interpreted relative to the directory
            containing the script running the statement.  For example, in a file called
    		"foo/bar.star", "bar.star" and "../foo/bar.star" would refer to that file.
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

def go_binary(workdir: string, target: string, cgo: bool = False):
    """
    Builds a Go binary for all platforms, by invoking "go build <target>" in workdir.

    Returns:
        A ReferenceList.
    """

def oci_layer(*inputs: ReferenceList):
    """
    Packages a filesystem into a OCI ("docker") layer.  Layers retain the platform annotation of their
    parent FS.

    Args:
        inputs: A list of inputs.  Each input becomes a new layer.
    Returns:
        A ReferenceList.
    """
