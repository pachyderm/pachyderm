# This is small script to setup a test repo in a new project (test_large_repo_project/test_large_repo).
# This test repo has a large number of files as defined by NUMBER_OF_FILES created in it. It is useful
# for testing how the Jupyterlab extension scales with any number of files.
#
# Once this script is run you should see a new project and repo created in the Pachyderm cluster with
# the url `test_large_repo_project/test_large_repo`. You can then use the Jupyterlab extension to mount
# this repo with the master branch and see how it behaves with a large number of files.
#
# There are some important behaviors that should be tested
# * While loading the files the text for each file should be grey, but each file can still be opened.
# * Changing directories while loading should stop the loading of the rest of the files. You can verify this by
#   opening up the developer tools and seeing that the requests for more files has stopped.
# * Changing directories back and forth to a large directory should not result in any files from the previous attempt appearing
#   twice.
# * While loading you should be able to use the scroll bar to navigate through files while loading. Since files are constantly
#   being added the same location in the scroll bar should not be the same file until finished loading.
# * While loading if you are not interacting with the scroll bar the files shown should not change.
# * Once all the files are loaded the text for all files should be black. Also the scroll bar should be able to navigate
#   to the same set of files when jumping to and from the same point in the scroll bar. You can verify this either by trying to click
#   on the same location in the scroll bar or just setting the scroll bar to exact values with the developer tools.
# * The local file cache should not hold on to any files once the directory is changed. You can verify this by checking the
#   MountDrive._cache in the developer tools.

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