""" This file patches methods onto the PPS Noun objects.
This is done, as opposed to subclassing them, for these methods
  to automatically be accessible on any objects included in the
  response from the API.
Note: These are internally patched and this file should
  not be imported directly by users.
"""
from . import CreatePipelineRequest, Service, Spout

def _Create_Pipeline_Request__post_init__(self: "CreatePipelineRequest") -> None:
    """
    Ensure that we serialize an empty Spout object
    """
    if self.spout and self.spout == Spout():
        self.spout = Spout(service=Service())

CreatePipelineRequest.__post_init__ = _Create_Pipeline_Request__post_init__
