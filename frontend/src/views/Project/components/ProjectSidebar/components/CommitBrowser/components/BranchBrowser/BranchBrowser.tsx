import {RepoQuery} from '@graphqlTypes';
import classNames from 'classnames';
import React, {useRef} from 'react';

import useIntersection from '@dash-frontend/hooks/useIntersection';
import {SearchableDropdown} from '@pachyderm/components';

import styles from './BranchBrowser.module.css';
import useBranchBrowser from './hooks/useBranchBrowser';

type BranchBrowserProps = {
  repo: RepoQuery['repo'];
  repoBaseRef: React.RefObject<HTMLDivElement>;
};

const BranchBrowser: React.FC<BranchBrowserProps> = ({repo, repoBaseRef}) => {
  const {handleBranchClick, dropdownItems, selectedBranch} = useBranchBrowser({
    branches: repo.branches,
  });

  const browserRef = useRef<HTMLDivElement>(null);
  const isStuck = useIntersection(browserRef.current, repoBaseRef.current);

  return (
    <div
      className={classNames(styles.base, {
        [styles.stuck]: isStuck,
      })}
      ref={browserRef}
    >
      <SearchableDropdown
        searchOpts={{placeholder: 'Search a branch by name'}}
        menuOpts={{className: styles.menu}}
        items={dropdownItems}
        onSelect={handleBranchClick}
        emptyResultsContent={'No matching branches found.'}
      >
        <strong>{`Branch: ${selectedBranch}`}</strong>
      </SearchableDropdown>
    </div>
  );
};

export default BranchBrowser;
