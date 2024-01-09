import React from 'react';

import {CaptionTextSmall, FullPagePanelModal} from '@pachyderm/components';

import BranchSelect from './components/BranchSelect';
import CommitList from './components/CommitList';
import useLeftPanel from './hooks/useLeftPanel';
import styles from './LeftPanel.module.css';

type LeftPanelProps = {
  selectedCommitId?: string;
};

const LeftPanel: React.FC<LeftPanelProps> = ({selectedCommitId}) => {
  const {commits, loading, page, setPage, hasNextPage, contentLength} =
    useLeftPanel(selectedCommitId);

  const breadCrumbText = `Commit: ${
    selectedCommitId && `${selectedCommitId.slice(0, 6)}...`
  }`;
  return (
    <FullPagePanelModal.LeftPanel>
      <div className={styles.base}>
        <div className={styles.branchSelect}>
          <div>
            <CaptionTextSmall>Branch</CaptionTextSmall>
            <BranchSelect />
          </div>
        </div>
        <div className={styles.breadCrumbs} data-testid={'LeftPanel_crumb'}>
          <CaptionTextSmall>{breadCrumbText}</CaptionTextSmall>
        </div>
        <div className={styles.data}>
          <CommitList
            selectedCommitId={selectedCommitId}
            commits={commits}
            loading={loading}
            page={page}
            setPage={setPage}
            hasNextPage={hasNextPage}
            contentLength={contentLength}
          />
        </div>
      </div>
    </FullPagePanelModal.LeftPanel>
  );
};

export default LeftPanel;
