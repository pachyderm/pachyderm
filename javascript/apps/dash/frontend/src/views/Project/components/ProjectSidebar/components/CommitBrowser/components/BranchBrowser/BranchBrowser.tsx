import {SearchableDropdown} from '@pachyderm/components';
import React from 'react';

import {RepoQuery} from '@graphqlTypes';

import styles from './BranchBrowser.module.css';
import useBranchBrowser from './hooks/useBranchBrowser';

type BranchBrowserProps = {
  repo?: RepoQuery['repo'];
};

const BranchBrowser: React.FC<BranchBrowserProps> = ({repo}) => {
  const {handleBranchClick, branchId, dropdownItems} = useBranchBrowser({
    branches: repo?.branches,
  });

  return (
    <SearchableDropdown
      selectedId={branchId}
      searchOpts={{placeholder: 'Search a branch by name'}}
      menuOpts={{className: styles.menu}}
      items={dropdownItems}
      onSelect={handleBranchClick}
      emptyResultsContent={'No matching branches found.'}
    >
      Commits (Branch: {branchId})
    </SearchableDropdown>
  );
};

export default BranchBrowser;
