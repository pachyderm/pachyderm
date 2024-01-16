import classnames from 'classnames';
import capitalize from 'lodash/capitalize';
import React from 'react';

import {CommitInfo} from '@dash-frontend/api/pfs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {FileCommitState} from '@dash-frontend/lib/types';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  CaptionTextSmall,
  Link,
  CaptionText,
  SkeletonDisplayText,
} from '@pachyderm/components';

import styles from './FileHistoryItem.module.css';
import useFileHistoryItem from './hooks/useFileHistoryItem';

type FileHistoryItemProps = {commit: CommitInfo};

const FileHistoryItem: React.FC<FileHistoryItemProps> = ({commit}) => {
  const {repoId, projectId, filePath, commitId} = useUrlState();

  const {loading, commitAction} = useFileHistoryItem(commit);

  return (
    <Link
      className={classnames(styles.listItem, styles.link, {
        [styles.selected]: commit.commit?.id === commitId,
      })}
      to={
        !loading && commitAction !== FileCommitState.DELETED
          ? fileBrowserRoute({
              projectId,
              repoId: repoId,
              commitId: commit.commit?.id || '',
              filePath,
            })
          : undefined
      }
    >
      <div className={styles.dateAndIdWrapper}>
        <div className={styles.dateAndStatus}>
          <span>{getStandardDateFromISOString(commit.started)}</span>

          {loading ? (
            <SkeletonDisplayText className={styles.skeletonDisplayText} />
          ) : (
            <CaptionTextSmall
              className={classnames({
                [styles.deleted]: commitAction === FileCommitState.DELETED,
              })}
            >
              {capitalize(commitAction || '-')}
            </CaptionTextSmall>
          )}
        </div>
        <div className={styles.textWrapper}>
          <CaptionText className={styles.commitId}>
            {commit.commit?.id}
          </CaptionText>
        </div>
      </div>
      {commit?.description && (
        <div className={styles.description}>{commit?.description}</div>
      )}
    </Link>
  );
};

export default FileHistoryItem;
