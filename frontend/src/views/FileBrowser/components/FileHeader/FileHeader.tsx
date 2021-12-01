import {
  Group,
  ArrowRightSVG,
  ArrowLeftSVG,
  Link,
  Search,
  Tooltip,
  InfoSVG,
  SkeletonDisplayText,
  Icon,
} from '@pachyderm/components';
import {format, fromUnixTime, formatDistanceStrict} from 'date-fns';
import findIndex from 'lodash/findIndex';
import React, {useMemo} from 'react';

import CommitPath from '@dash-frontend/components/CommitPath';
import Header from '@dash-frontend/components/Header';
import useCommits, {COMMIT_LIMIT} from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import useFileBrowser from '../../hooks/useFileBrowser';

import styles from './FileHeader.module.css';

type FileHeaderProps = {
  fileFilter: string;
  setFileFilter: React.Dispatch<React.SetStateAction<string>>;
};

const FileHeader: React.FC<FileHeaderProps> = ({fileFilter, setFileFilter}) => {
  const {repoId, commitId, branchId, projectId, filePath} = useUrlState();
  const {fileToPreview, loading} = useFileBrowser();

  const {commits, loading: commitsLoading} = useCommits({
    projectId,
    repoName: repoId,
    branchName: branchId,
    number: COMMIT_LIMIT,
  });

  const {previousCommit, nextCommit} = useMemo(() => {
    const reposInBranch = commits || [];
    const index = findIndex(reposInBranch, (commit) => commit.id === commitId);

    const previousCommit =
      reposInBranch.length > index + 1 ? reposInBranch[index + 1].id : null;
    const nextCommit = index > 0 ? reposInBranch[index - 1].id : null;

    return {previousCommit, nextCommit};
  }, [commitId, commits]);

  const currentCommit = useMemo(() => {
    return (commits || []).find((commit) => commit.id === commitId);
  }, [commitId, commits]);

  const tooltipInfo = currentCommit ? (
    <>
      Size: {currentCommit.sizeDisplay}
      <br />
      {currentCommit.started && currentCommit.finished ? (
        <>
          Duration:{' '}
          {formatDistanceStrict(
            fromUnixTime(currentCommit.started),
            fromUnixTime(currentCommit.finished),
          )}
        </>
      ) : null}
      <br />
      Origin: {currentCommit.originKind}
      {currentCommit.description && <br />}
      {currentCommit.description}
    </>
  ) : (
    <span />
  );

  return (
    <Header appearance="light" hasSubheader>
      <Group spacing={8} align="center" className={styles.header}>
        <h4 className={styles.commit}>
          <CommitPath repo={repoId} branch={branchId} commit={commitId} />
        </h4>
        {!loading && !fileToPreview && (
          <Search
            value={fileFilter}
            className={styles.search}
            placeholder="Filter current folder by name"
            onSearch={setFileFilter}
            aria-label="search files"
          />
        )}
        <Group spacing={32}>
          {previousCommit && (
            <Tooltip
              tooltipText="See Previous Commit"
              placement="bottom"
              tooltipKey="See Previous Commit"
            >
              <Link
                className={styles.navigate}
                to={fileBrowserRoute({
                  repoId,
                  branchId,
                  projectId,
                  commitId: previousCommit,
                  filePath: filePath || undefined,
                })}
              >
                <ArrowLeftSVG className={styles.directionLeft} />
                Older
              </Link>
            </Tooltip>
          )}
          {nextCommit && (
            <Tooltip
              tooltipText="See Next Commit"
              placement="bottom"
              tooltipKey="See Next Commit"
            >
              <Link
                className={styles.navigate}
                to={fileBrowserRoute({
                  repoId,
                  branchId,
                  projectId,
                  commitId: nextCommit,
                  filePath: filePath || undefined,
                })}
              >
                Newer
                <ArrowRightSVG className={styles.directionRight} />
              </Link>
            </Tooltip>
          )}
        </Group>
      </Group>
      <Group spacing={16} className={styles.subHeader}>
        <span className={styles.dateField}>
          {commitsLoading ? (
            <SkeletonDisplayText />
          ) : (
            currentCommit?.started && (
              <>
                <span className={styles.datelabel}>Started:</span>
                {` `}
                {format(
                  fromUnixTime(currentCommit.started),
                  'MM/dd/yyyy h:mm:ssaaa',
                )}
              </>
            )
          )}
        </span>
        <span className={styles.dateField}>
          {commitsLoading ? (
            <SkeletonDisplayText />
          ) : (
            currentCommit?.finished && (
              <span>
                <span className={styles.datelabel}>Ended: </span>
                {` `}
                {format(
                  fromUnixTime(currentCommit.finished),
                  'MM/dd/yyyy h:mm:ssaaa',
                )}
              </span>
            )
          )}
        </span>
        <Tooltip
          tooltipText={tooltipInfo}
          size="extraLarge"
          placement="bottom"
          tooltipKey="See Next Commit"
        >
          <span className={styles.datelabel}>
            More Info
            <Icon small className={styles.infoIcon}>
              <InfoSVG />
            </Icon>
          </span>
        </Tooltip>
      </Group>
    </Header>
  );
};

export default FileHeader;
