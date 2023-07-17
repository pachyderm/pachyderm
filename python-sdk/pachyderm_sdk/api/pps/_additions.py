from . import CreatePipelineRequest, Service, Spout

def _Create_Pipeline_Request__post_init__(self: "CreatePipelineRequest") -> None:
    """
    Ensure that we serialize an empty Spout object
    """
    if self.spout and self.spout == Spout():
        self.spout = Spout(service=Service())

CreatePipelineRequest.__post_init__ = _Create_Pipeline_Request__post_init__
