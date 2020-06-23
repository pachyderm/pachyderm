import json
import asyncio
import collections

RunResult = collections.namedtuple("RunResult", ["rc", "stdout", "stderr"])

client_version = None
async def get_client_version():
    """Gets the pachctl client version"""
    global client_version
    if client_version is None:
        client_version = (await capture("pachctl", "version", "--client-only")).strip()
    return client_version

class RedactedString(str):
    pass

async def run(cmd, *args, raise_on_error=True, stdin=None, capture_output=False, timeout=None):
    """Runs a command asynchronously"""

    print_args = [cmd, *[a if not isinstance(a, RedactedString) else "[redacted]" for a in args]]
    print_status("running: `{}`".format(" ".join(print_args)))

    proc = await asyncio.create_subprocess_exec(
        cmd, *args,
        stdin=asyncio.subprocess.PIPE if stdin is not None else None,
        stdout=asyncio.subprocess.PIPE if capture_output else None,
        stderr=asyncio.subprocess.PIPE if capture_output else None,
    )
    
    if timeout is None:
        stdin = stdin.encode("utf8") if stdin is not None else None
        result = await proc.communicate(input=stdin)
    else:
        result = await asyncio.wait_for(proc.communicate(input=stdin), timeout=timeout)

    if capture_output:
        stdout, stderr = result
        stdout = stdout.decode("utf8")
        stderr = stderr.decode("utf8")
    else:
        stdout, stderr = None, None

    if raise_on_error and proc.returncode:
        raise Exception(f"unexpected return code from `{cmd}`: {proc.returncode}")

    return RunResult(rc=proc.returncode, stdout=stdout, stderr=stderr)

async def capture(cmd, *args, **kwargs):
    _, stdout, _ = await run(cmd, *args, capture_output=True, **kwargs)
    return stdout

def find_in_json(j, f):
    """
    Recurses through a JSON value, looking for a value according to the input
    function `f`.
    """

    if f(j):
        return j

    iter = None
    if isinstance(j, dict):
        iter = j.values()
    elif isinstance(j, list):
        iter = j

    if iter is not None:
        for sub_j in iter:
            v = find_in_json(sub_j, f)
            if v is not None:
                return v

def print_status(status):
    """Prints a status message"""
    print(f"===> {status}")

async def retry(f, attempts=10, sleep=1.0):
    """
    Repeatedly retries an operation, ignore exceptions, n times with a given
    sleep between runs.
    """
    count = 0
    while count < attempts:
        try:
            return await f()
        except:
            count += 1
            if count >= attempts:
                raise
            await asyncio.sleep(sleep)

async def ping():
    await run("pachctl", "version", timeout=5)
