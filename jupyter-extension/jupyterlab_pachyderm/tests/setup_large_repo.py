# This is small script to setup a test repo in a new project (test_large_repo_project/test_large_repo).
# This test repo has a large number of files as defined by NUMBER_OF_FILES created in it. It is useful
# for testing how the Jupyterlab extension scales with any number of files.

import math

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs

NUMBER_OF_FILES = 20000
FILES_PER_COMMIT = 10

def main(): 
    client = Client.from_config()
    
    # Optionally call this if you want a clean slate
    # client.pfs.delete_all()

    project = pfs.Project(name="test_large_repo_project")
    repo = pfs.Repo(name="test_large_repo", project=project)
    branch = pfs.Branch.from_uri(f"{repo}@master")


    # Clean previous run
    try:
        client.pfs.delete_repo(repo=repo) 
    except:
        print('project already deleted')
    try:
        client.pfs.delete_project(project=project) 
    except:
        print('project already deleted')
    
    # Setup new project and repo
    client.pfs.create_project(project=project)
    client.pfs.create_repo(repo=repo)

    for b in range(math.floor(NUMBER_OF_FILES / FILES_PER_COMMIT)):
        with client.pfs.commit(branch=branch) as commit:
            print(f'starting commit {commit}')
            for i in range(FILES_PER_COMMIT):
                file_number = i + (b * FILES_PER_COMMIT)
                path = f"/hello{file_number}.py"
                print(f'Writing path={path}')
                commit.put_file_from_bytes(path=path, data=b"print('hello')")
        commit.wait()

if __name__=="__main__": 
    main() 