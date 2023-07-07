# Contributing Guide

## Getting Started
We use [poetry](https://python-poetry.org/) to both manage dependencies
  and publish this package. You can find information about installing this
  software [here](https://python-poetry.org/docs/).

Once both `poetry` is installed, you can use `poetry` to create your
  virtual environment for this project. The following command
  (run at the root of this project) will create a virtual environment
  within the `.venv` directory:
```bash
poetry install
source .venv/bin/activate
```

## Code

### Layout
Code layout, as of Jun. 30, 2023:
```
.
├── docs - Auto-generated API docs
├── examples - The examples
├── pachderm_sdk/
│   ├── __main__.py - Location for package scripts (cli)
│   ├── api/ - Generated API code. `extension.py` files are hand-written.
│   ├── client.py - The higher-level `Client` class.
│   ├── config.py - Config file parsing code.
│   ├── constants.py - Centralized location for constant values.
│   ├── datum_batching.py - Decorator for datum-batching code.
│   ├── errors.py - Centralized location for custom errors.
│   └── interceptor.py - gRPC interceptor implementation.
├── proto/ - Code generation of the protobuf files.
├── tests/ - Pytests
│   ├── fixtures.py - Location for pytest fixtures used throughout tests.
├── Dockerfile.datum-batching-test - Dockerfile used to build image for datum-batching test.
├── poetry.lock - Lock file for this packages dependencies.
└── pyproject.toml - Package specification file.
```

## Testing
To execute the tests, run the following command:
```bash
poetry run pytest -vvv tests
```

Setting the following environment variables is needed for all the tests
to run properly:
* PACH_PYTHON_TEST_HOST - Hostname to your pachyderm cluster
* PACH_PYTHON_TEST_PORT - Port over which to connect to you pachyderm cluster
* PACH_PYTHON_TEST_PORT_ENTERPRISE - Port over which to connect to your enterprise-enabled cluster
* PACH_PYTHON_AUTH_TOKEN - Auth token for your enterprise cluster (found in .pachyderm/config.json)

### Formatting

This project uses the [black](https://github.com/psf/black) code formatter.
Pytest will automatically check that your code changes pass the black formatter.
To run the black formatter:
```bash
black <file-or-directory>
```

## Releasing
Releasing the package is done through CircleCI.
In the event you need to cut a manual release, you need to set the proper package version
  in the pyproject.toml file and then run the following command:
```bash
poetry publish --build
```
The above command will prompt the user for a PyPI username/password.