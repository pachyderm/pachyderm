import React, {useCallback, useState} from 'react';

import {FileInfo} from '@dash-frontend/api/pfs';
import EmptyState from '@dash-frontend/components/EmptyState';
import {CaptionTextSmall, LoadingDots} from '@pachyderm/components';

import BranchSelect from './components/BranchSelect';
import CommitSelect from './components/CommitSelect';
import FileSearch from './components/FileSearch';
import FileShortcuts from './components/FileShortcuts';
import DirectoryContents from './components/FileTree';
import styles from './LeftPanel.module.css';

type LeftPanelProps = {
  selectedCommitId?: string;
  isCommitOpen: boolean;
};

const LeftPanel: React.FC<LeftPanelProps> = ({
  selectedCommitId,
  isCommitOpen,
}) => {
  const [selectedFile, setSelectedFile] = useState<FileInfo>();
  const onClickFile = useCallback((file: FileInfo) => {
    setSelectedFile(file);
  }, []);

  return (
    <div className={styles.base}>
      <div className={styles.select}>
        <div>
          <CaptionTextSmall>Branch</CaptionTextSmall>
          <BranchSelect buttonType="input" />
        </div>
      </div>
      <div className={styles.select}>
        <div>
          <CaptionTextSmall>Commit</CaptionTextSmall>
          <CommitSelect selectedCommitId={selectedCommitId} />
        </div>
      </div>

      {!isCommitOpen && (
        <div className={styles.search}>
          <FileSearch
            selectedCommitId={selectedCommitId}
            onClickFile={onClickFile}
          />
          <FileShortcuts />
        </div>
      )}
      <section className={styles.tree} aria-label="File Tree">
        {isCommitOpen ? (
          <>
            <EmptyState title="This commit is currently open" />
          </>
        ) : selectedCommitId ? (
          <DirectoryContents
            selectedCommitId={selectedCommitId}
            path="/"
            open
            initial
            key={selectedFile?.file?.path}
          />
        ) : (
          <LoadingDots />
        )}
      </section>
    </div>
  );
};

export default LeftPanel;
