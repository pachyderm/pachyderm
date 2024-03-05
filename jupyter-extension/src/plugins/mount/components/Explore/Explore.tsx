import React from 'react';
import {showErrorMessage} from '@jupyterlab/apputils';

import {Repo, Mount, ListMountsResponse} from '../../types';
import {DropdownCombobox} from '../../../../components/DropdownCombobox/DropdownCombobox';
import {unmountAll, mount, getMountedStatus} from '../../api';

type ExploreProps = {
  mounted: Mount[];
  unmounted: Repo[];
  updateData: (response: ListMountsResponse) => void;
};

const Explore: React.FC<ExploreProps> = ({mounted, unmounted, updateData}) => {
  // Avoids rendering the dropdowns until mount information is loaded.
  if (mounted.length === 0 && unmounted.length === 0) {
    return <></>;
  }

  // In the event of some how multiple repos being mounted we should unmount all to reset to the default state.
  if (mounted.length > 1) {
    unmountAll(updateData);
    showErrorMessage(
      'Unexpected Error',
      'Multiple repos have been mounted somehow so all repos have been unmounted.',
    );
    return <></>;
  }

  const {projectRepos, selectedProjectRepo, branches, selectedBranch} =
    getMountedStatus(mounted, unmounted);

  return (
    <div className="pachyderm-explore-view">
      <DropdownCombobox
        testIdPrefix="ProjectRepo-"
        initialSelectedItem={selectedProjectRepo}
        items={projectRepos}
        placeholder="project/repo"
        onSelectedItemChange={(projectRepo) => {
          if (!projectRepo) {
            unmountAll(updateData);
            return;
          }

          mount(updateData, projectRepo);
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
            if (!selectedBranch) {
              return;
            }

            mount(updateData, selectedProjectRepo, selectedBranch);
          }}
        />
      )}
    </div>
  );
};

export default Explore;
