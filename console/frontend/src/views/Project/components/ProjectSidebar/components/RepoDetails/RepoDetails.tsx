import React from 'react';
import {Route, Switch} from 'react-router-dom';

import {BrandedEmptyIcon} from '@dash-frontend/components/BrandedIcon';
import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import ContentBox from '@dash-frontend/components/ContentBox';
import {ContentBoxContent} from '@dash-frontend/components/ContentBox/ContentBox';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState/EmptyState';
import GlobalIdCopy from '@dash-frontend/components/GlobalIdCopy';
import GlobalIdDisclaimer from '@dash-frontend/components/GlobalIdDisclaimer';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {InputOutputNodesMap} from '@dash-frontend/lib/types';
import {
  LINEAGE_REPO_PATH,
  LINEAGE_FILE_BROWSER_PATH,
  LINEAGE_FILE_BROWSER_PATH_LATEST,
  PROJECT_FILE_BROWSER_PATH_LATEST,
  PROJECT_FILE_BROWSER_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {fileUploadRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  SkeletonDisplayText,
  ButtonLink,
  useModal,
  Group,
  Button,
  Link,
  Icon,
  UploadSVG,
  Tabs,
} from '@pachyderm/components';

import Title from '../Title';

import CommitDetails from './components/CommitDetails';
import CommitList from './components/CommitList';
import RepoInfo from './components/RepoInfo';
import UserMetadata from './components/UserMetadata';
import useRepoDetails from './hooks/useRepoDetails';
import styles from './RepoDetails.module.css';

export enum TAB_ID {
  OVERVIEW = 'overview',
  INFO = 'info',
  METADATA = 'metadata',
}

type RepoDetailsProps = {
  pipelineOutputsMap?: InputOutputNodesMap;
};

const RepoDetails: React.FC<RepoDetailsProps> = ({pipelineOutputsMap = {}}) => {
  const {
    repo,
    commit,
    givenCommitId,
    commitBranches,
    repoError,
    currentRepoLoading,
    commitDiff,
    diffLoading,
    projectId,
    repoId,
    hasRepoEditRoles,
    getPathToFileBrowser,
    hasRepoRead,
    hasRepoWrite,
    globalId,
  } = useRepoDetails();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  if (!currentRepoLoading && repoError) {
    return (
      <div className={styles.emptyRepoMessage}>
        <BrandedEmptyIcon className={styles.emptyIcon} disableDefaultStyling />
        <h5>Unable to load Repo and commit data</h5>
        <p>
          We weren&apos;t able to fetch any information about this repo,
          including its latest commit. Please try refreshing this page. If this
          issue keeps happening, contact our customer team.
        </p>
      </div>
    );
  }

  const contentDetails: ContentBoxContent[] = [
    {
      label: 'Commit ID',
      value: commit ? <GlobalIdCopy id={commit.commit?.id || ''} /> : 'N/A',
      responsiveValue: commit ? (
        <GlobalIdCopy id={commit.commit?.id || ''} shortenId />
      ) : (
        'N/A'
      ),
      dataTestId: 'RepoDetails__commit_id',
    },
  ];
  const fileBrowserContentDetails: ContentBoxContent[] = [];
  if (commitBranches && commitBranches.length > 0) {
    fileBrowserContentDetails.push({
      label: 'Branch',
      value: commitBranches.map((b) => b.branch?.name || '').join(', '),
      dataTestId: 'RepoDetails__commit_branch',
    });
  }
  fileBrowserContentDetails.push(
    {
      label: 'Commit Type',
      value: commit?.origin?.kind?.toLowerCase() || '-',
      dataTestId: 'RepoDetails__commit_type',
    },
    {
      label: 'Start',
      value: commit ? getStandardDateFromISOString(commit.started) : 'N/A',
      dataTestId: 'RepoDetails__commit_start',
    },
  );
  contentDetails.push(...fileBrowserContentDetails);

  return (
    <div className={styles.base} data-testid="RepoDetails__base">
      <BrandedTitle title="Repo" />
      <div className={styles.titleSection}>
        {currentRepoLoading ? (
          <SkeletonDisplayText data-testid="RepoDetails__repoNameSkeleton" />
        ) : (
          <Title>{repo?.repo?.name}</Title>
        )}
        <Switch>
          <Route path={LINEAGE_REPO_PATH} exact>
            {repo?.authInfo?.roles && (
              <Description loading={currentRepoLoading} term="Your Roles">
                <Group spacing={8}>
                  {repo?.authInfo?.roles.join(', ') || 'None'}
                  <ButtonLink onClick={openRolesModal}>
                    {hasRepoEditRoles ? 'Set Roles' : 'See All Roles'}
                  </ButtonLink>
                </Group>
              </Description>
            )}
            {repo?.authInfo?.roles && rolesModalOpen && (
              <RepoRolesModal
                show={rolesModalOpen}
                onHide={closeRolesModal}
                projectName={projectId}
                repoName={repoId}
                readOnly={!hasRepoEditRoles}
              />
            )}
          </Route>
        </Switch>
      </div>

      <Tabs initialActiveTabId={TAB_ID.OVERVIEW}>
        <Tabs.TabsHeader className={styles.tabsHeader}>
          <div className={styles.tabs}>
            <Tabs.Tab id={TAB_ID.OVERVIEW}>Overview</Tabs.Tab>
            <Tabs.Tab id={TAB_ID.INFO}>Repo Info</Tabs.Tab>
            {hasRepoRead && (
              <Tabs.Tab id={TAB_ID.METADATA}>User Metadata</Tabs.Tab>
            )}
          </div>
        </Tabs.TabsHeader>
        <Tabs.TabPanel id={TAB_ID.OVERVIEW}>
          <section className={styles.section}>
            <Switch>
              <Route path={LINEAGE_REPO_PATH} exact>
                {commit?.commit?.id && (
                  <div className={styles.commitSummaryLabel}>
                    Commit Summary
                    {commit && (
                      <Button
                        buttonType="ghost"
                        to={getPathToFileBrowser({
                          projectId,
                          repoId,
                          commitId: commit.commit?.id,
                        })}
                        disabled={!commit}
                        aria-label="Inspect Commit"
                      >
                        Inspect Commit
                      </Button>
                    )}
                  </div>
                )}
                {(currentRepoLoading || commit) && (
                  <GlobalIdDisclaimer
                    globalId={globalId}
                    artifact="commit"
                    startTime={commit?.started}
                  />
                )}
                {(currentRepoLoading || commit) && (
                  <ContentBox
                    loading={currentRepoLoading}
                    content={contentDetails}
                  />
                )}
              </Route>
            </Switch>

            <Switch>
              <Route
                path={[
                  LINEAGE_FILE_BROWSER_PATH,
                  LINEAGE_FILE_BROWSER_PATH_LATEST,
                  PROJECT_FILE_BROWSER_PATH,
                  PROJECT_FILE_BROWSER_PATH_LATEST,
                ]}
              >
                {(currentRepoLoading || commit) && (
                  <ContentBox
                    loading={currentRepoLoading}
                    content={fileBrowserContentDetails}
                  />
                )}
              </Route>
            </Switch>

            {(currentRepoLoading || commit) && commit?.description && (
              <ContentBox
                loading={currentRepoLoading}
                content={[
                  {
                    label: 'Commit Message',
                    value: commit.description,
                    dataTestId: 'RepoDetails__commit_message',
                  },
                ]}
              />
            )}

            {!currentRepoLoading && hasRepoRead && !commit && (
              <>
                <EmptyState
                  title={<>This repo doesn&apos;t have any data</>}
                  message={
                    <>
                      This is normal for new repositories. If you are interested
                      in learning more:
                    </>
                  }
                  linkToDocs={{
                    text: 'View our documentation about managing data',
                    pathWithoutDomain: '/prepare-data/ingest-data/',
                  }}
                />
                {hasRepoWrite && (
                  <Link
                    className={styles.link}
                    to={fileUploadRoute({projectId, repoId})}
                  >
                    Upload files
                    <Icon small color="inherit">
                      <UploadSVG className={styles.linkIcon} />
                    </Icon>
                  </Link>
                )}
              </>
            )}
            {!currentRepoLoading && !hasRepoRead && (
              <EmptyState
                title={<>{`You don't have permission to view this repo`}</>}
                noAccess
                message={
                  <>
                    {`You'll need a role of repoReader or higher to view commit data
        about this repo.`}
                  </>
                }
                linkToDocs={{
                  text: 'Read more about authorization',
                  pathWithoutDomain: '/set-up/authorization/',
                }}
              />
            )}
          </section>

          {commit?.commit?.id && (
            <CommitDetails
              commit={commit}
              diffLoading={diffLoading}
              commitDiff={commitDiff?.diff}
            />
          )}

          <Route path={LINEAGE_REPO_PATH} exact>
            {repo && commit && !globalId && <CommitList repo={repo} />}
          </Route>
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.INFO}>
          <section className={styles.section}>
            <RepoInfo
              pipelineOutputsMap={pipelineOutputsMap}
              repo={repo}
              repoError={repoError}
              currentRepoLoading={currentRepoLoading}
            />
          </section>
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.METADATA}>
          <section className={styles.section}>
            <UserMetadata
              metadataType="repo"
              metadata={repo?.metadata}
              editable={hasRepoWrite}
            />
            {commit?.commit?.id && (
              <UserMetadata
                metadataType="commit"
                metadata={commit?.metadata}
                id={commit?.commit?.id}
                editable={hasRepoWrite}
                captionText={
                  givenCommitId
                    ? `Commit ${givenCommitId?.slice(0, 6)}...`
                    : 'For the most recent commit'
                }
              />
            )}
          </section>
        </Tabs.TabPanel>
      </Tabs>
    </div>
  );
};

export default RepoDetails;
