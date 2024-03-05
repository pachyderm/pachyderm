import {Repo, Mount, ListMountsResponse} from './types';
import {requestAPI} from '../../handler';

export const unmountAll = async (
  updateData?: (response: ListMountsResponse) => void,
): Promise<void> => {
  const response = await requestAPI<ListMountsResponse>('_unmount_all', 'PUT');
  if (updateData) {
    return updateData(response);
  }

  return;
};

export const mount = async (
  updateData: (response: ListMountsResponse) => void,
  projectRepo: string,
  branch = 'master',
): Promise<void> => {
  const [project, repo] = projectRepo.split('/');

  await unmountAll();
  const response = await requestAPI<ListMountsResponse>('_mount', 'PUT', {
    mounts: [
      {
        name:
          branch === 'master'
            ? `${project}_${repo}`
            : `${project}_${repo}_${branch}`,
        repo: repo,
        branch: branch,
        project: project,
        mode: 'ro',
      },
    ],
  });

  return updateData(response);
};

type MountedStatus = {
  projectRepos: string[];
  selectedProjectRepo: string | null;
  branches: string[] | null;
  selectedBranch: string | null;
};

export const getMountedStatus = (
  mounted: Mount[],
  unmounted: Repo[],
): MountedStatus => {
  const projectRepos: string[] = [];
  const projectRepoToBranches: {[projectRepo: string]: string[]} = {};
  for (const repo of unmounted) {
    const projectRepoKey = `${repo.project}/${repo.repo}`;
    projectRepoToBranches[projectRepoKey] = repo.branches;
    projectRepos.push(projectRepoKey);
  }

  let selectedProjectRepo: string | null = null;
  let selectedBranch: string | null = null;
  let branches: string[] | null = null;
  if (mounted.length === 1) {
    const mountedBranch = mounted[0];
    selectedProjectRepo = `${mountedBranch.project}/${mountedBranch.repo}`;
    selectedBranch = mountedBranch.branch;
    if (!projectRepos.includes(selectedProjectRepo)) {
      projectRepos.push(selectedProjectRepo);
      projectRepoToBranches[selectedProjectRepo] = [];
    }
    if (
      !projectRepoToBranches[selectedProjectRepo].includes(mountedBranch.branch)
    ) {
      projectRepoToBranches[selectedProjectRepo].push(mountedBranch.branch);
    }
    branches = projectRepoToBranches[selectedProjectRepo];
    branches?.sort();
  }
  projectRepos.sort();

  return {projectRepos, selectedProjectRepo, branches, selectedBranch};
};
