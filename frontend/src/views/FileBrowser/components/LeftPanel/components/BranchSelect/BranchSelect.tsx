import React, {useMemo, useState} from 'react';
import {useHistory} from 'react-router';

import {useBranches} from '@dash-frontend/hooks/useBranches';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserLatestRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  CloseSVG,
  Dropdown,
  Icon,
  SearchSVG,
} from '@pachyderm/components';

import styles from './BranchSelect.module.css';

const BranchSelect: React.FC = () => {
  const browserHistory = useHistory();
  const {projectId, repoId, branchId} = useUrlState();
  const [searchFilter, setSearchFilter] = useState('');

  const {branches} = useBranches({
    projectId,
    repoId,
  });

  const updateSelectedBranch = (selectedBranchId?: string) => {
    setSearchFilter('');
    browserHistory.push(
      fileBrowserLatestRoute({
        projectId,
        repoId,
        branchId: selectedBranchId || 'default',
      }),
    );
  };

  const filteredBranches = useMemo(
    () =>
      branches && searchFilter
        ? branches.filter(
            (branch) => branch?.branch?.name?.indexOf(searchFilter) !== -1,
          )
        : branches,
    [branches, searchFilter],
  );

  return (
    <Dropdown className={styles.base}>
      <Dropdown.Button>{branchId}</Dropdown.Button>
      <Dropdown.Menu className={styles.menu}>
        <div className={styles.search}>
          <Icon className={styles.searchIcon} small>
            <SearchSVG aria-hidden />
          </Icon>

          <input
            className={styles.input}
            placeholder="Enter branch name"
            onChange={(e) => setSearchFilter(e.target.value)}
            value={searchFilter}
          />

          {searchFilter && (
            <Button
              aria-label="Clear"
              buttonType="ghost"
              onClick={() => setSearchFilter('')}
              IconSVG={CloseSVG}
            />
          )}
        </div>

        {(filteredBranches || []).map((value) => (
          <Dropdown.MenuItem
            closeOnClick
            onClick={() => updateSelectedBranch(value?.branch?.name)}
            key={value?.branch?.name}
            id={value?.branch?.name || 'default'}
          >
            {value?.branch?.name}
          </Dropdown.MenuItem>
        ))}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default BranchSelect;
