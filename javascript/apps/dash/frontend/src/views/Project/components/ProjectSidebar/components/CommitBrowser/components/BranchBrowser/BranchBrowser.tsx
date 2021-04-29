import {ButtonLink, ChevronDownSVG, SearchSVG} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useCallback, useMemo, useState} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {RepoQuery} from '@graphqlTypes';

import styles from './BranchBrowser.module.css';
import useBranchBrowser from './hooks/useBranchBrowser';

type BranchBrowserProps = {
  repo?: RepoQuery['repo'];
};

const BranchBrowser: React.FC<BranchBrowserProps> = ({repo}) => {
  const {branchId} = useUrlState();
  const [query, setQuery] = useState('');
  const branches = useMemo(() => {
    const branches = repo?.branches ? [...repo.branches] : [];

    branches.sort((a, b) => {
      if (a.name === 'master' || b.name > a.name) return -1;
      if (a.name > b.name) return 1;

      return 0;
    });
    branches.splice(1, 0, {id: 'none', name: 'none'});

    return query
      ? branches.filter((branch) => branch.name.includes(query))
      : branches;
  }, [repo, query]);
  const handleSearch = useCallback(
    (evt: React.ChangeEvent<HTMLInputElement>) => setQuery(evt.target.value),
    [],
  );
  const {
    browserRef,
    handleBranchClick,
    handleHeaderClick,
    isMenuOpen,
  } = useBranchBrowser();

  return (
    <div className={styles.base} ref={browserRef}>
      <h2 className={styles.header}>
        Commits (Branch: {branchId})
        <ButtonLink onClick={handleHeaderClick}>
          <ChevronDownSVG
            className={classNames(
              styles.chevron,
              isMenuOpen && styles.menuOpen,
            )}
          />
        </ButtonLink>
      </h2>
      {isMenuOpen && (
        <div className={styles.branchWrapper}>
          <div className={styles.search}>
            <SearchSVG />
            <input
              className={styles.searchInput}
              onChange={handleSearch}
              placeholder="Find a branch"
            />
          </div>
          <div className={styles.divider} />
          {branches.map((branch) => (
            <div
              className={classNames(
                styles.branch,
                branch.id === branchId && styles.currentBranch,
              )}
              key={branch.id}
              onClick={() => handleBranchClick(branch.id)}
            >
              {branch.name}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default BranchBrowser;
