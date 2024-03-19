import {getMountedStatus, getDefaultBranch} from '../api';

describe('getMountedStatus', () => {
  it('return status for a mounted branch in a repo with multiple branches', () => {
    const {
      projectRepos,
      selectedProjectRepo,
      branches,
      selectedBranch,
      projectRepoToBranches,
    } = getMountedStatus(
      [
        {
          name: 'name',
          repo: 'repo',
          project: 'project',
          branch: 'branch1',
        },
      ],
      [
        {
          repo: 'zzz',
          project: 'project',
          authorization: 'read',
          branches: ['ignoredBranch1', 'ignoredBranch2'],
        },
        {
          repo: 'repo',
          project: 'project',
          authorization: 'read',
          branches: ['zzz', 'branch2'],
        },
      ],
    );

    expect(projectRepos).toEqual(['project/repo', 'project/zzz']);
    expect(selectedProjectRepo).toBe('project/repo');
    expect(branches).toEqual(['branch1', 'branch2', 'zzz']);
    expect(selectedBranch).toBe('branch1');
    expect(projectRepoToBranches).toEqual({
      'project/zzz': ['ignoredBranch1', 'ignoredBranch2'],
      'project/repo': ['branch1', 'branch2', 'zzz'],
    });
  });

  it('return status for a mounted branch in a repo with one branch', () => {
    const {
      projectRepos,
      selectedProjectRepo,
      branches,
      selectedBranch,
      projectRepoToBranches,
    } = getMountedStatus(
      [
        {
          name: 'name',
          repo: 'repo',
          project: 'project',
          branch: 'branch1',
        },
      ],
      [
        {
          repo: 'repo2',
          project: 'project',
          authorization: 'read',
          branches: ['ignoredBranch1', 'ignoredBranch2'],
        },
      ],
    );

    expect(projectRepos).toEqual(['project/repo', 'project/repo2']);
    expect(selectedProjectRepo).toBe('project/repo');
    expect(branches).toEqual(['branch1']);
    expect(selectedBranch).toBe('branch1');
    expect(projectRepoToBranches).toEqual({
      'project/repo': ['branch1'],
      'project/repo2': ['ignoredBranch1', 'ignoredBranch2'],
    });
  });

  it('return status with no mounted branch', () => {
    const {
      projectRepos,
      selectedProjectRepo,
      branches,
      selectedBranch,
      projectRepoToBranches,
    } = getMountedStatus(
      [],
      [
        {
          repo: 'repo',
          project: 'project',
          authorization: 'read',
          branches: ['branch1', 'branch2'],
        },
        {
          repo: 'repo2',
          project: 'project',
          authorization: 'read',
          branches: ['ignoredBranch1', 'ignoredBranch2'],
        },
      ],
    );

    expect(projectRepos).toEqual(['project/repo', 'project/repo2']);
    expect(selectedProjectRepo).toBeNull();
    expect(branches).toBeNull();
    expect(selectedBranch).toBeNull();
    expect(projectRepoToBranches).toEqual({
      'project/repo': ['branch1', 'branch2'],
      'project/repo2': ['ignoredBranch1', 'ignoredBranch2'],
    });
  });
});

describe('getDefaultBranch', () => {
  it('return null if branches is empty', () => {
    expect(getDefaultBranch([])).toBeNull();
  });

  it('return master if master is included in the branches', () => {
    expect(getDefaultBranch(['branchzz', 'master', 'branchaa'])).toBe('master');
  });

  it('return choose the first branch after sorting', () => {
    expect(getDefaultBranch(['branchzz', 'branchaa'])).toBe('branchaa');
  });
});
