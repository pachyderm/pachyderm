import {getMountedStatus} from '../api';

describe('getMountedStatus', () => {
  it('return status for a mounted branch in a repo with multiple branches', () => {
    const {projectRepos, selectedProjectRepo, branches, selectedBranch} =
      getMountedStatus(
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
  });

  it('return status for a mounted branch in a repo with one branch', () => {
    const {projectRepos, selectedProjectRepo, branches, selectedBranch} =
      getMountedStatus(
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
  });

  it('return status with no mounted branch', () => {
    const {projectRepos, selectedProjectRepo, branches, selectedBranch} =
      getMountedStatus(
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
  });
});
