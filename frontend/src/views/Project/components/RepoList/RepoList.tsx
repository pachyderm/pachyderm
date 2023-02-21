import React, {useState} from 'react';

import TableView, {TableViewSection} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PROJECT_REPOS_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import CommitsList from './components/CommitsList';
import ReposTable from './components/ReposTable';

const TAB_IDS = {repos: 'repos', commits: 'commits'};

const RepoList: React.FC = () => {
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const {viewState} = useUrlQueryState();

  return (
    <TableView
      title="Repositories"
      noun="repo"
      tabsBasePath={PROJECT_REPOS_PATH}
      tabs={TAB_IDS}
      selectedItems={viewState.selectedRepos || []}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
      singleRowSelection
    >
      <TableViewSection id={TAB_IDS.repos}>
        <ReposTable filtersExpanded={filtersExpanded} />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.commits}>
        <CommitsList
          selectedRepo={
            viewState.selectedRepos ? viewState.selectedRepos[0] : ''
          }
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
    </TableView>
  );
};

export default RepoList;
