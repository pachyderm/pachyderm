from pathlib import Path

DEFAULT_PROJECT = "default"
TEST_NOTEBOOK = Path(__file__).parent.joinpath('data/TestNotebook.ipynb')
TEST_REQUIREMENTS = TEST_NOTEBOOK.with_name("requirements.txt")
