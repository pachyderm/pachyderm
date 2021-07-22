import {SearchableDropdown} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useRef} from 'react';

import useIntersection from '@dash-frontend/hooks/useIntersection';
import {RepoQuery} from '@graphqlTypes';

import styles from './BranchBrowser.module.css';
import useBranchBrowser from './hooks/useBranchBrowser';

type BranchBrowserProps = {
  repo?: RepoQuery['repo'];
  repoBaseRef: React.RefObject<HTMLDivElement>;
};

const BranchBrowser: React.FC<BranchBrowserProps> = ({repo, repoBaseRef}) => {
  const {handleBranchClick, branchId, dropdownItems} = useBranchBrowser({
    branches: repo?.branches,
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
        Commits (Branch: {branchId})
      </SearchableDropdown>
    </div>
  );
};

export default BranchBrowser;
