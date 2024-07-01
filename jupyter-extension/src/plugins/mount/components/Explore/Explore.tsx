import React from 'react';
import {showErrorMessage} from '@jupyterlab/apputils';

import {Repo, Repos, MountedRepo, Branch} from '../../types';
import {DropdownCombobox} from '../../../../utils/components/DropdownCombobox/DropdownCombobox';
import {TextInput} from '../../../../utils/components/TextInput/TextInput';

type ExploreProps = {
  repos: Repos;
  mountedRepo: MountedRepo | null;
  updateMountedRepo: (
    repo: Repo | null,
    mountedBranch: Branch | null,
    commit: string | null,
  ) => void;
  mountedRepoUri: string | null;
};

const Explore: React.FC<ExploreProps> = ({
  repos,
  mountedRepo,
  updateMountedRepo,
  mountedRepoUri,
}) => {
  // Avoids rendering the dropdowns until mount information is loaded.
  if (!mountedRepo && Object.keys(repos).length === 0) {
    return <></>;
  }

  return (
    <div className="pachyderm-explore-view">
      <DropdownCombobox
        testIdPrefix="ProjectRepo-"
        initialSelectedItem={mountedRepo?.repo.uri}
        items={Object.keys(repos)}
        placeholder="project/repo"
        onSelectedItemChange={(repoUri, selectItem) => {
          (async () => {
            if (!repoUri) {
              updateMountedRepo(null, null, null);
              return;
            }

            const repo = repos[repoUri];
            if (repo.branches.length === 0) {
              updateMountedRepo(null, null, null);
              showErrorMessage(
                'No Branches',
                `${repo.name} has no branches to mount`,
              );
              selectItem(null);
              return;
            }

            updateMountedRepo(repo, null, null);
          })();
        }}
      />
      {!mountedRepo ? (
        <></>
      ) : (
        <>
          <DropdownCombobox
            testIdPrefix="Branch-"
            initialSelectedItem={mountedRepo.branch?.name || ''}
            items={mountedRepo.repo.branches.map((branch) => branch.name)}
            placeholder="branch"
            onSelectedItemChange={(mountedBranchName) => {
              (async () => {
                // When clicking the dropdown that causes the selected item to be cleared. We should do nothing in that case.
                if (!mountedBranchName) {
                  return;
                }

                let mountedBranch: Branch | null = null;
                for (const branch of mountedRepo.repo.branches) {
                  if (branch.name === mountedBranchName) {
                    mountedBranch = branch;
                    break;
                  }
                }

                if (!mountedBranch) {
                  showErrorMessage(
                    'Explore View Reset',
                    'A branch that was removed was selected. Resetting explore view.',
                  );
                  updateMountedRepo(null, null, null);
                  return;
                }

                updateMountedRepo(mountedRepo.repo, mountedBranch, null);
              })();
            }}
          />
          <TextInput
            onSubmit={(value) =>
              updateMountedRepo(mountedRepo.repo, null, value)
            }
            placeholder="Commit"
            testIdPrefix="Commit-"
          />
        </>
      )}
      <p className="pachyderm-Explore-mount-uri">
        Mounted Uri: {mountedRepoUri}
      </p>
    </div>
  );
};

export default Explore;
