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
      initialSelectId={branchId}
      searchOpts={{placeholder: 'Search a branch by name', autoComplete: 'off'}}
      menuOpts={{className: styles.menu}}
      storeSelected
      items={dropdownItems}
      onSelect={handleBranchClick}
      emptyResultsContent={'No matching branches found.'}
    >
      Commits (Branch: {branchId})
    </SearchableDropdown>
  );
};

export default BranchBrowser;
