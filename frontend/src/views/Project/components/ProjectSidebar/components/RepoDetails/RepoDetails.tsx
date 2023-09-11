import objectHash from 'object-hash';
import React, {useMemo} from 'react';
import {Route, Switch} from 'react-router-dom';

import {BrandedEmptyIcon} from '@dash-frontend/components/BrandedIcon';
import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState/EmptyState';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import {PipelineLink} from '@dash-frontend/components/ResourceLink';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {InputOutputNodesMap} from '@dash-frontend/lib/types';
import {LINEAGE_REPO_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  SkeletonDisplayText,
  CaptionTextSmall,
  ButtonLink,
  useModal,
  Group,
  Button,
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
    editRolesPermission,
    getPathToFileBrowser,
  } = useRepoDetails();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const pipelineOutputs = useMemo(() => {
    const repoNodeName = objectHash({
      project: repo?.projectId,
      name: repo?.name,
    });

    return pipelineOutputsMap[repoNodeName] || [];
  }, [pipelineOutputsMap, repo?.name, repo?.projectId]);

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
          <Title>{repo?.name}</Title>
        )}
        {repo?.description && (
          <div className={styles.description}>{repo?.description}</div>
        )}
        <Switch>
          <Route path={LINEAGE_REPO_PATH} exact>
            {repo?.authInfo?.rolesList && (
              <Description loading={currentRepoLoading} term="Your Roles">
                <Group spacing={8}>
                  {repo?.authInfo?.rolesList.join(', ') || 'None'}
                  <ButtonLink onClick={openRolesModal}>
                    {editRolesPermission ? 'Set Roles' : 'See All Roles'}
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
          {repo ? getStandardDate(repo.createdAt) : 'N/A'}
        </Description>
        <Switch>
          <Route path={LINEAGE_REPO_PATH} exact>
            {(currentRepoLoading || commit) && (
              <Description
                loading={currentRepoLoading}
                term="Most Recent Commit Start"
              >
                {commit ? getStandardDate(commit.started) : 'N/A'}
              </Description>
            )}
            {(currentRepoLoading || commit) && (
              <Description
                loading={currentRepoLoading}
                term="Most Recent Commit ID"
              >
                {commit ? (
                  <CommitIdCopy commit={commit.id} clickable small longId />
                ) : (
                  'N/A'
                )}
              </Description>
            )}

            {repo?.authInfo?.rolesList && rolesModalOpen && (
              <RepoRolesModal
                show={rolesModalOpen}
                onHide={closeRolesModal}
                projectName={projectId}
                repoName={repoId}
                readOnly={!editRolesPermission}
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

      {!currentRepoLoading && (!commit || repo?.branches.length === 0) && (
        <EmptyState
          title={<>This repo doesn&apos;t have any branches</>}
          message={
            <>
              This is normal for new repositories, but we still wanted to notify
              you because Pachyderm didn&apos;t detect a branch on our end.
            </>
          }
          linkToDocs={{
            text: 'Try creating a branch and pushing a commit.',
            pathWithoutDomain: 'concepts/data-concepts/branch/',
          }}
        />
      )}

      {commit?.id && (
        <>
          <CaptionTextSmall className={styles.commitDetailsLabel}>
            <Switch>
              <Route path={LINEAGE_REPO_PATH} exact>
                Current Commit Stats
                {commit && (
                  <Button
                    buttonType="ghost"
                    to={getPathToFileBrowser({
                      projectId,
                      repoId,
                      commitId: commit.id,
                      branchId: commit.branch?.name || '',
                    })}
                    disabled={!commit}
                    aria-label="Inspect Current Commit"
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
            commitDiff={commitDiff?.commitDiff}
            repo={repo}
          />
        </>
      )}

      <Route path={LINEAGE_REPO_PATH} exact>
        {repo && commit && <CommitList repo={repo} />}
      </Route>
    </div>
  );
};

export default RepoDetails;
