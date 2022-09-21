import {SkeletonDisplayText, Tabs} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import capitalize from 'lodash/capitalize';
import React, {useRef, useCallback} from 'react';
import {Helmet} from 'react-helmet';
import {useRouteMatch} from 'react-router';

import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import {PipelineLink, EgressLink} from '@dash-frontend/components/ResourceLink';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  LINEAGE_PATH,
  LINEAGE_REPO_PATH,
  PROJECT_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

import CommitBrowser from '../CommitBrowser';
import CommitDetails from '../CommitBrowser/components/CommitDetails';
import Title from '../Title';

import {TAB_ID} from './constants/tabIds';
import styles from './RepoDetails.module.css';

type RepoDetailsProps = {
  dagLinks?: Record<string, string[]>;
  dagsLoading?: boolean;
};

const noBranchRepoMessage = 'There are no branches on this repo!';

const RepoDetails: React.FC<RepoDetailsProps> = ({dagsLoading, dagLinks}) => {
  const {loading, repo} = useCurrentRepo();
  const {repoId} = useUrlState();

  const currentRepoLoading = loading || repoId !== repo?.id;

  const {viewState} = useUrlQueryState();
  const repoBaseRef = useRef<HTMLDivElement>(null);
  const filterEgressLink = useCallback(
    (egress: boolean) => (name: string) => {
      try {
        const url = new URL(name);
        return egress === Boolean(url);
      } catch {
        return egress === false;
      }
    },
    [],
  );

  const inputs = (dagLinks && dagLinks[`${repo?.name}_repo`]) || [];

  const egressOutputs = inputs.filter(filterEgressLink(true));
  const pipelineOutputs = inputs.filter(filterEgressLink(false));

  const lineageMatch = useRouteMatch({
    path: LINEAGE_PATH,
  });

  const tabsBasePath = lineageMatch ? LINEAGE_REPO_PATH : PROJECT_REPO_PATH;

  const BranchBrowser = () =>
    repo?.branches?.length ? (
      <CommitBrowser repo={repo} repoBaseRef={repoBaseRef} />
    ) : (
      <EmptyState title={LETS_START_TITLE} message={noBranchRepoMessage} />
    );

  return (
    <div className={styles.base} ref={repoBaseRef}>
      <Helmet>
        <title>Repo - Pachyderm Console</title>
      </Helmet>
      <div className={styles.title}>
        {currentRepoLoading ? (
          <SkeletonDisplayText data-testid="RepoDetails__repoNameSkeleton" />
        ) : (
          <Title>{repo?.name}</Title>
        )}
      </div>

      <Tabs.RouterTabs basePathTabId={TAB_ID.COMMITS} basePath={tabsBasePath}>
        <Tabs.TabsHeader className={styles.tabsHeader}>
          {Object.values(TAB_ID).map((tabId) => {
            let text = capitalize(tabId);
            if (viewState.globalIdFilter && text === 'Commits') {
              text = 'Commit';
            }
            return (
              <Tabs.Tab id={tabId} key={tabId}>
                {text}
              </Tabs.Tab>
            );
          })}
        </Tabs.TabsHeader>

        <Tabs.TabPanel id={TAB_ID.INFO}>
          <dl className={styles.repoInfo}>
            {repo?.linkedPipeline && (
              <Description loading={currentRepoLoading} term="Linked Pipeline">
                <div className={styles.pipelineGroup}>
                  <PipelineLink name={repo?.linkedPipeline?.id} />
                </div>
              </Description>
            )}
            {pipelineOutputs.length > 0 && (
              <Description loading={currentRepoLoading} term="Inputs To">
                <div className={styles.pipelineGroup}>
                  {pipelineOutputs.map((name) => (
                    <PipelineLink name={name} key={name} />
                  ))}
                </div>
              </Description>
            )}
            {egressOutputs.length > 0 && (
              <Description loading={currentRepoLoading} term="Egress To">
                <div className={styles.pipelineGroup}>
                  {egressOutputs.map((name) => (
                    <EgressLink name={name} key={name} />
                  ))}
                </div>
              </Description>
            )}
            <Description loading={currentRepoLoading} term="Created">
              {repo ? format(fromUnixTime(repo.createdAt), 'MM/d/yyyy') : 'N/A'}
            </Description>
            <Description loading={currentRepoLoading} term="Description">
              {repo?.description ? repo.description : 'N/A'}
            </Description>
            <Description term="Size">
              {!currentRepoLoading ? (
                repo?.sizeDisplay
              ) : (
                <SkeletonDisplayText />
              )}
            </Description>
          </dl>
        </Tabs.TabPanel>
        <div className={styles.commitsTab}>
          <Tabs.TabPanel id={TAB_ID.COMMITS}>
            {viewState.globalIdFilter
              ? !dagsLoading && (
                  <CommitDetails
                    commitId={viewState.globalIdFilter}
                    repo={repo}
                  />
                )
              : !currentRepoLoading && <BranchBrowser />}
          </Tabs.TabPanel>
        </div>
      </Tabs.RouterTabs>
    </div>
  );
};

export default RepoDetails;
