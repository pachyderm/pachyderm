import React from 'react';

import SortableList from '../SortableList/SortableList';

import {Repo, Mount, ProjectInfo, ListMountsResponse} from '../../types';

type ExploreProps = {
  mounted: Mount[];
  unmounted: Repo[];
  projects: ProjectInfo[];
  openPFS: (path: string) => void;
  updateData: (data: ListMountsResponse) => void;
};

const Explore: React.FC<ExploreProps> = ({
  mounted,
  unmounted,
  projects,
  openPFS,
  updateData,
}) => {
  return (
    <div className="pachyderm-explore-view">
      <div className="pachyderm-mount-react-wrapper">
        <div className="pachyderm-mount-base">
          <SortableList
            open={openPFS}
            items={mounted}
            updateData={updateData}
            mountedItems={[]}
            type={'mounted'}
            projects={[]}
          />
        </div>
      </div>
      <div className="pachyderm-mount-react-wrapper">
        <div className="pachyderm-mount-base">
          <div className="pachyderm-mount-base-title">
            Unmounted Repositories
          </div>
          <SortableList
            open={openPFS}
            items={unmounted}
            updateData={updateData}
            mountedItems={mounted}
            type={'unmounted'}
            projects={projects}
          />
        </div>
      </div>
    </div>
  );
};

export default Explore;
