# This is small script to setup a test repo in a new project (test_large_repo_project/test_large_repo).
# This test repo has a large number of files as defined by NUMBER_OF_FILES created in it. It is useful
# for testing how the Jupyterlab extension scales with any number of files.

from pachyderm_sdk import Client
from pachyderm_sdk.api import pfs

NUMBER_OF_FILES = 20000

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

    with client.pfs.commit(branch=branch) as c:
        print(f'starting commit {c}')
        for i in range(NUMBER_OF_FILES):
            path = f"/hello{i}.py"
            print(f'Writing path={path}')
            c.put_file_from_bytes(path=path, data=b"print('hello')")

if __name__=="__main__": 
    main() 