import objectHash from 'object-hash';
import React, {useMemo} from 'react';
import {Route, Switch} from 'react-router-dom';

import {BrandedEmptyIcon} from '@dash-frontend/components/BrandedIcon';
import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState/EmptyState';
import ExpandableText from '@dash-frontend/components/ExpandableText';
import GlobalIdCopy from '@dash-frontend/components/GlobalIdCopy';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import {PipelineLink} from '@dash-frontend/components/ResourceLink';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {InputOutputNodesMap} from '@dash-frontend/lib/types';
import {LINEAGE_REPO_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {fileUploadRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  SkeletonDisplayText,
  CaptionTextSmall,
  ButtonLink,
  useModal,
  Group,
  Button,
  Link,
  Icon,
  UploadSVG,
} from '@pachyderm/components';

import Title from '../Title';

import CommitDetails from './components/CommitDetails';
import CommitList from './components/CommitList';
import useRepoDetails from './hooks/useRepoDetails';
import styles from './RepoDetails.module.css';

type RepoDetailsProps = {
  pipelineOutputsMap?: InputOutputNodesMap;
};

const RepoDetails: React.FC<RepoDetailsProps> = ({pipelineOutputsMap = {}}) => {
  const {
    repo,
    commit,
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

  const pipelineOutputs = useMemo(() => {
    const repoNodeName = objectHash({
      project: repo?.repo?.project?.name,
      name: repo?.repo?.name,
    });

    return pipelineOutputsMap[repoNodeName] || [];
  }, [pipelineOutputsMap, repo?.repo?.name, repo?.repo?.project?.name]);

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

  return (
    <div className={styles.base} data-testid="RepoDetails__base">
      <BrandedTitle title="Repo" />
      <div className={styles.titleSection}>
        {currentRepoLoading ? (
          <SkeletonDisplayText data-testid="RepoDetails__repoNameSkeleton" />
        ) : (
          <Title>{repo?.repo?.name}</Title>
        )}
        {repo?.description && (
          <div className={styles.description}>{repo?.description}</div>
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
          </Route>
        </Switch>
        {pipelineOutputs.length > 0 && (
          <Description loading={currentRepoLoading} term="Inputs To">
            {pipelineOutputs.map(({name}) => (
              <PipelineLink name={name} key={name} />
            ))}
          </Description>
        )}
        <Description
          loading={currentRepoLoading}
          term="Repo Created"
          error={repoError}
        >
          {repo ? getStandardDateFromISOString(repo.created) : 'N/A'}
        </Description>
        <Switch>
          <Route path={LINEAGE_REPO_PATH} exact>
            {(currentRepoLoading || commit) && (
              <Description
                loading={currentRepoLoading}
                term={`${
                  !globalId
                    ? 'Most Recent Commit Start'
                    : 'Global ID Commit Start'
                }`}
              >
                {commit ? getStandardDateFromISOString(commit.started) : 'N/A'}
              </Description>
            )}
            {(currentRepoLoading || commit) && commit?.description && (
              <Description
                loading={currentRepoLoading}
                term={
                  !globalId
                    ? 'Most Recent Commit Message'
                    : 'Global ID Commit Message'
                }
              >
                <ExpandableText text={commit.description} />
              </Description>
            )}
            {(currentRepoLoading || commit) && (
              <Description
                loading={currentRepoLoading}
                term={!globalId ? 'Most Recent Commit ID' : 'Global ID'}
              >
                {commit ? <GlobalIdCopy id={commit.commit?.id || ''} /> : 'N/A'}
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

          <Route>
            {commit?.description && (
              <Description term="Selected Commit Description">
                {commit.description}
              </Description>
            )}
          </Route>
        </Switch>
      </div>

      {!currentRepoLoading && hasRepoRead && !commit && (
        <>
          <EmptyState
            title={<>This repo doesn&apos;t have any data</>}
            message={
              <>
                This is normal for new repositories. If you are interested in
                learning more:
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

      {commit?.commit?.id && (
        <>
          <CaptionTextSmall className={styles.commitDetailsLabel}>
            <Switch>
              <Route path={LINEAGE_REPO_PATH} exact>
                {!globalId ? 'Current Commit Stats' : 'Global ID Commit Stats'}
                {commit && (
                  <Button
                    buttonType="ghost"
                    to={getPathToFileBrowser({
                      projectId,
                      repoId,
                      commitId: commit.commit.id,
                    })}
                    disabled={!commit}
                    aria-label="Inspect Commit"
                  >
                    Inspect Commit
                  </Button>
                )}
              </Route>
              <Route>Selected Commit Stats</Route>
            </Switch>
          </CaptionTextSmall>
          <CommitDetails
            commit={commit}
            diffLoading={diffLoading}
            commitDiff={commitDiff?.diff}
          />
        </>
      )}

      <Route path={LINEAGE_REPO_PATH} exact>
        {repo && commit && !globalId && <CommitList repo={repo} />}
      </Route>
    </div>
  );
};

export default RepoDetails;
