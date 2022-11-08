import {format, fromUnixTime, formatDistanceStrict} from 'date-fns';
import findIndex from 'lodash/findIndex';
import React, {useMemo} from 'react';

import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import Header from '@dash-frontend/components/Header';
import useCommits, {COMMIT_LIMIT} from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
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
  Switch,
} from '@pachyderm/components';

import useFileBrowser from '../../hooks/useFileBrowser';

import styles from './FileHeader.module.css';

type FileHeaderProps = {
  fileFilter: string;
  setFileFilter: React.Dispatch<React.SetStateAction<string>>;
  diffOnly: boolean;
  setDiffOnly: (diff: boolean) => void;
};

const FileHeader: React.FC<FileHeaderProps> = ({
  fileFilter,
  setFileFilter,
  setDiffOnly,
  diffOnly,
}) => {
  const {repoId, commitId, branchId, projectId, filePath} = useUrlState();
  const {fileToPreview, loading, files} = useFileBrowser();

  const {commits, loading: commitsLoading} = useCommits({
    args: {
      projectId,
      repoName: repoId,
      branchName: branchId,
      number: COMMIT_LIMIT,
    },
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

  const diffString = useMemo(() => {
    const updates = [];
    let updatesString = '';

    if (files?.diff?.filesAdded && files?.diff.filesAdded > 0)
      updates.push({count: files.diff.filesAdded, status: 'added'});
    if (files?.diff?.filesUpdated && files?.diff.filesUpdated > 0)
      updates.push({count: files.diff.filesUpdated, status: 'updated'});
    if (files?.diff?.filesDeleted && files?.diff.filesDeleted > 0)
      updates.push({count: files.diff.filesDeleted, status: 'deleted'});

    if (updates[0])
      updatesString = `${updates[0].count} ${
        updates[0].count > 1 ? 'Files' : 'File'
      } ${updates[0].status}`;
    if (updates[1])
      updatesString.concat(`, ${updates[1].count}
           ${updates[1].status}`);
    if (updates[2])
      updatesString.concat(`, ${updates[2].count}
           ${updates[2].status}`);

    return updatesString;
  }, [files]);

  const tooltipInfo = currentCommit ? (
    <>
      Size: {currentCommit.sizeDisplay}
      <br />
      {currentCommit.started !== -1 && currentCommit.finished !== -1 ? (
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
        <CommitIdCopy repo={repoId} branch={branchId} commit={commitId} />
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
      <Group className={styles.subHeader}>
        <Group spacing={32}>
          {diffString && (
            <span className={styles.dateField}>
              {loading ? (
                <SkeletonDisplayText />
              ) : (
                files &&
                files.diff && (
                  <>
                    <b>{diffString}</b>
                    {files.diff.size !== 0 &&
                      ` (${files.diff.size > 0 ? '+' : ''}${
                        files.diff.sizeDisplay
                      })`}
                  </>
                )
              )}
            </span>
          )}
          <span className={styles.dateField}>
            {commitsLoading ? (
              <SkeletonDisplayText />
            ) : (
              currentCommit?.started && (
                <>
                  <b>Started:</b>
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
                  <b>Ended:</b>
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
              <b>More Info</b>
              <Icon small className={styles.infoIcon}>
                <InfoSVG />
              </Icon>
            </span>
          </Tooltip>
        </Group>
        <div className={styles.switchItem}>
          <Switch
            className={styles.switch}
            defaultChecked={diffOnly}
            onChange={() => setDiffOnly(!diffOnly)}
            aria-label="Show diff only"
          />
          <span>Show diff only</span>
        </div>
      </Group>
    </Header>
  );
};

export default FileHeader;
