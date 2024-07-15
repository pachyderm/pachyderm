import React, {useMemo, useState} from 'react';
import {useHistory} from 'react-router';

import {useBranches} from '@dash-frontend/hooks/useBranches';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
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

type BranchSelectProps = {
  buttonType?: React.ComponentProps<typeof Dropdown.Button>['buttonType'];
};

const BranchSelect = ({buttonType}: BranchSelectProps) => {
  const browserHistory = useHistory();
  const {projectId, repoId} = useUrlState();
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();

  const [searchFilter, setSearchFilter] = useState('');

  const {branches} = useBranches({
    projectId,
    repoId,
  });

  const updateSelectedBranch = (selectedBranchId?: string) => {
    setSearchFilter('');

    const params = {
      branchId: selectedBranchId ? selectedBranchId : undefined,
    };

    browserHistory.push(
      `${fileBrowserLatestRoute(
        {
          projectId,
          repoId,
        },
        false,
      )}?${getUpdatedSearchParams(params, false)}`,
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
      <Dropdown.Button buttonType={buttonType}>
        {searchParams.branchId || 'Viewing all commits'}
      </Dropdown.Button>
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

        {searchParams.branchId && (
          <Dropdown.MenuItem
            closeOnClick
            onClick={() => updateSelectedBranch(undefined)}
            key="AllCommits"
            id="AllCommits"
          >
            View all commits
          </Dropdown.MenuItem>
        )}

        {(filteredBranches || []).map((value) => {
          if (!value?.branch?.name) {
            return null;
          }

          return (
            <Dropdown.MenuItem
              closeOnClick
              onClick={() => updateSelectedBranch(value?.branch?.name)}
              key={value?.branch?.name}
              id={value?.branch?.name}
            >
              {value?.branch?.name}
            </Dropdown.MenuItem>
          );
        })}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default BranchSelect;
