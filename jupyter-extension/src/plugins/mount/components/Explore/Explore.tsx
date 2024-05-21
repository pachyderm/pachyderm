import React, {useEffect} from 'react';
import {showErrorMessage} from '@jupyterlab/apputils';

import {Repo, Mount, ListMountsResponse} from '../../types';
import {DropdownCombobox} from '../../../../utils/components/DropdownCombobox/DropdownCombobox';
import {unmountAll, mount, getMountedStatus, getDefaultBranch} from '../../api';

type ExploreProps = {
  mounted: Mount[];
  unmounted: Repo[];
  updateData: (response: ListMountsResponse) => void;
  changeDirectory: (directory: string) => Promise<void>;
};

const Explore: React.FC<ExploreProps> = ({
  mounted,
  unmounted,
  updateData,
  changeDirectory,
}) => {
  useEffect(() => {
    if (mounted.length === 1) {
      changeDirectory(`/${mounted[0].name}`);
    }
  }, [mounted.length]);

  // Avoids rendering the dropdowns until mount information is loaded.
  if (mounted.length === 0 && unmounted.length === 0) {
    return <></>;
  }

  // In the event of some how multiple repos being mounted we should unmount all to reset to the default state.
  if (mounted.length > 1) {
    unmountAll().then((response) => updateData(response));
    showErrorMessage(
      'Unexpected Error',
      'Multiple repos have been mounted somehow so all repos have been unmounted.',
    );
    return <></>;
  }

  const {
    projectRepos,
    selectedProjectRepo,
    branches,
    selectedBranch,
    projectRepoToBranches,
  } = getMountedStatus(mounted, unmounted);

  return (
    <div className="pachyderm-explore-view">
      <DropdownCombobox
        testIdPrefix="ProjectRepo-"
        initialSelectedItem={selectedProjectRepo}
        items={projectRepos}
        placeholder="project/repo"
        onSelectedItemChange={(projectRepo, selectItem) => {
          (async () => {
            if (!projectRepo) {
              updateData(await unmountAll());
              await changeDirectory('/');
              return;
            }

            const defaultBranch = getDefaultBranch(
              projectRepoToBranches[projectRepo],
            );
            if (!defaultBranch) {
              showErrorMessage(
                'No Branches',
                `${projectRepo} has no branches to mount`,
              );
              selectItem(null);
              return;
            }

            const response = await mount(projectRepo, defaultBranch);
            updateData(response);
            await changeDirectory(`/${response.mounted[0].name}`);
          })();
        }}
      />
      {!branches || !selectedProjectRepo ? (
        <></>
      ) : (
        <DropdownCombobox
          testIdPrefix="Branch-"
          initialSelectedItem={selectedBranch}
          items={branches}
          placeholder="branch"
          onSelectedItemChange={(selectedBranch) => {
            (async () => {
              if (!selectedBranch) {
                return;
              }

              const response = await mount(selectedProjectRepo, selectedBranch);
              updateData(response);
              await changeDirectory(`/${response.mounted[0].name}`);
            })();
          }}
        />
      )}
    </div>
  );
};

export default Explore;
