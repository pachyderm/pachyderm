import React from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  CaptionText,
  CopySVG,
  Tooltip,
  useClipboardCopy,
  Link,
} from '@pachyderm/components';

import styles from './FileHeader.module.css';

type FileHeaderProps = {
  commitId?: string;
};

const FileHeader: React.FC<FileHeaderProps> = ({commitId}) => {
  const {filePath, projectId, repoId} = useUrlState();

  const commitPath = commitId
    ? fileBrowserRoute({
        projectId,
        repoId,
        commitId,
      })
    : '';

  const {copy} = useClipboardCopy(
    `${repoId}@${commitId}${filePath ? `:${filePath}` : ''}` || '',
  );

  return (
    <div className={styles.base}>
      <div className={styles.path} data-testid="FileHeader__path">
        <CaptionText color="black">{repoId}@</CaptionText>
        {filePath ? (
          <Link inline to={commitPath}>
            <CaptionText className={styles.link}>
              {commitId && commitId.slice(0, 6)}
            </CaptionText>
          </Link>
        ) : (
          <CaptionText color="black">{commitId}</CaptionText>
        )}
        {filePath && (
          <>
            <CaptionText color="black">:{filePath}</CaptionText>
          </>
        )}
      </div>
      <Tooltip tooltipText="Copy commit id">
        <Button
          IconSVG={CopySVG}
          buttonType="ghost"
          color="black"
          onClick={copy}
          aria-label="Copy commit id"
        />
      </Tooltip>
    </div>
  );
};

export default FileHeader;
