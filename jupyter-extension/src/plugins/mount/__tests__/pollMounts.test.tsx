import {PollMounts} from '../pollMounts';
import {MountedRepo, Repo, Branch} from '../types';
import * as handler from '../../../handler';

jest.mock('../../../handler');

describe('PollMounts', () => {
  const mockedRequestAPI = handler.requestAPI as jest.MockedFunction<
    typeof handler.requestAPI
  >;
  beforeEach(() => {
    mockedRequestAPI.mockClear();
    localStorage.clear();
  });

  it('should get mounted repo from localStorage', () => {
    const expectedMountedRepo: MountedRepo = {
      branch: {
        name: 'branch',
        uri: 'test_repo@branch',
      },
      repo: {
        name: 'name',
        project: 'test',
        uri: 'test_repo',
        branches: [
          {
            name: 'branch',
            uri: 'test_repo@branch',
          },
        ],
      },
      commit: null,
    };
    localStorage.setItem(
      PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
      JSON.stringify(expectedMountedRepo),
    );

    const pollMounts = new PollMounts('testPollMounts');

    expect(pollMounts.mountedRepo).toEqual(expectedMountedRepo);
    expect(mockedRequestAPI).toHaveBeenCalledWith('explore/mount', 'PUT', {
      commit_uri: expectedMountedRepo.branch?.uri,
    });
  });

  it('should clear mounted repo from local storage if data is not JSON', () => {
    localStorage.setItem(
      PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
      '{{{{{ def not json }}}}}',
    );

    const pollMounts = new PollMounts('testPollMounts');

    expect(pollMounts.mountedRepo).toBeNull();
    expect(
      localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
    ).toBeNull();
    expect(mockedRequestAPI).toHaveBeenCalledWith('explore/unmount', 'PUT');
  });

  describe('updateMountedRepo()', () => {
    it('should default to master if repo exists and mountedBranch is null', async () => {
      const pollMounts = new PollMounts('testPollMounts');
      const masterBranch: Branch = {
        name: 'master',
        uri: 'test_repo@master',
      };
      const repo: Repo = {
        name: 'name',
        project: 'test',
        uri: 'test_repo',
        branches: [
          {
            name: 'branch',
            uri: 'test_repo@branch',
          },
          masterBranch,
        ],
      };
      const expectedMountedRepo: MountedRepo = {
        branch: masterBranch,
        repo,
        commit: null,
      };

      await pollMounts.updateMountedRepo(repo, null, null);

      expect(pollMounts.mountedRepo).toEqual(expectedMountedRepo);
      expect(
        localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
      ).toEqual(JSON.stringify(expectedMountedRepo));
      expect(mockedRequestAPI).toHaveBeenCalledWith('explore/mount', 'PUT', {
        commit_uri: expectedMountedRepo.branch?.uri,
      });
    });

    it('should default to the first branch if repo exists and mountedBranch is null and master branch does not exist', async () => {
      const pollMounts = new PollMounts('testPollMounts');
      const firstBranch: Branch = {
        name: 'branch1',
        uri: 'test_repo@branch1',
      };
      const repo: Repo = {
        name: 'name',
        project: 'test',
        uri: 'test_repo',
        branches: [
          firstBranch,
          {
            name: 'branch2',
            uri: 'test_repo@branch2',
          },
        ],
      };
      const expectedMountedRepo: MountedRepo = {
        branch: firstBranch,
        repo,
        commit: null,
      };

      await pollMounts.updateMountedRepo(repo, null, null);

      expect(pollMounts.mountedRepo).toEqual(expectedMountedRepo);
      expect(
        localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
      ).toEqual(JSON.stringify(expectedMountedRepo));
      expect(mockedRequestAPI).toHaveBeenCalledWith('explore/mount', 'PUT', {
        commit_uri: expectedMountedRepo.branch?.uri,
      });
    });

    it('should clear mounted repo when passed mount repo is null', async () => {
      const pollMounts = new PollMounts('testPollMounts');
      localStorage.setItem(
        PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
        '{{{{{ def not json }}}}}',
      );

      await pollMounts.updateMountedRepo(null, null, null);

      expect(pollMounts.mountedRepo).toBeNull();
      expect(
        localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
      ).toBeNull();
      expect(mockedRequestAPI).not.toHaveBeenCalled();
    });

    it('should set mounted repo if repo and mountedBranch exist', async () => {
      const pollMounts = new PollMounts('testPollMounts');
      localStorage.setItem(
        PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
        '{{{{{ def not json }}}}}',
      );
      const mountedBranch: Branch = {
        name: 'branch2',
        uri: 'test_repo@branch2',
      };
      const repo: Repo = {
        name: 'name',
        project: 'test',
        uri: 'test_repo',
        branches: [
          {
            name: 'branch1',
            uri: 'test_repo@branch1',
          },
          mountedBranch,
        ],
      };
      const expectedMountedRepo: MountedRepo = {
        branch: mountedBranch,
        repo,
        commit: null,
      };

      await pollMounts.updateMountedRepo(repo, mountedBranch, null);

      expect(pollMounts.mountedRepo).toEqual(expectedMountedRepo);
      expect(
        localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
      ).toEqual(JSON.stringify(expectedMountedRepo));
      expect(mockedRequestAPI).toHaveBeenCalledWith('explore/mount', 'PUT', {
        commit_uri: expectedMountedRepo.branch?.uri,
      });
    });

    it('should set mounted repo with commit', async () => {
      const pollMounts = new PollMounts('testPollMounts');
      localStorage.setItem(
        PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
        '{{{{{ def not json }}}}}',
      );
      const mountedBranch: Branch = {
        name: 'branch2',
        uri: 'test_repo@branch2',
      };
      const repo: Repo = {
        name: 'name',
        project: 'test',
        uri: 'test_repo',
        branches: [
          {
            name: 'branch1',
            uri: 'test_repo@branch1',
          },
          mountedBranch,
        ],
      };
      const commit = '17ad1a170664403a98cb75db3e20fa11';
      const expectedMountedRepo: MountedRepo = {
        branch: mountedBranch,
        repo,
        commit,
      };

      await pollMounts.updateMountedRepo(repo, mountedBranch, commit);

      expect(pollMounts.mountedRepo).toEqual(expectedMountedRepo);
      expect(
        localStorage.getItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY),
      ).toEqual(JSON.stringify(expectedMountedRepo));
      expect(mockedRequestAPI).toHaveBeenCalledWith('explore/mount', 'PUT', {
        commit_uri: 'test_repo@branch2=17ad1a170664403a98cb75db3e20fa11',
      });
    });
  });
});
