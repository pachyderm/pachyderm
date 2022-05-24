import {SkeletonDisplayText, Tabs} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import capitalize from 'lodash/capitalize';
import React, {useRef} from 'react';
import {Helmet} from 'react-helmet';
import {useRouteMatch} from 'react-router';

import Description from '@dash-frontend/components/Description';
import RepoPipelineLink from '@dash-frontend/components/RepoPipelineLink';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
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

const RepoDetails: React.FC<RepoDetailsProps> = ({dagsLoading, dagLinks}) => {
  const {loading, repo} = useCurrentRepo();
  const {viewState} = useUrlQueryState();
  const repoBaseRef = useRef<HTMLDivElement>(null);

  const lineageMatch = useRouteMatch({
    path: LINEAGE_PATH,
  });

  const tabsBasePath = lineageMatch ? LINEAGE_REPO_PATH : PROJECT_REPO_PATH;

  return (
    <div className={styles.base} ref={repoBaseRef}>
      <Helmet>
        <title>Repo - Pachyderm Console</title>
      </Helmet>
      <div className={styles.title}>
        {loading ? (
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
              <Description loading={loading} term="Linked Pipeline">
                <div className={styles.pipelineGroup}>
                  <RepoPipelineLink
                    name={repo?.linkedPipeline?.id}
                    type="pipeline"
                  />
                </div>
              </Description>
            )}
            {dagLinks && dagLinks[`${repo?.name}_repo`] && (
              <Description loading={loading} term="Inputs To">
                <div className={styles.pipelineGroup}>
                  {dagLinks[`${repo?.name}_repo`].map((name) => (
                    <RepoPipelineLink name={name} type="pipeline" key={name} />
                  ))}
                </div>
              </Description>
            )}
            <Description loading={loading} term="Created">
              {repo ? format(fromUnixTime(repo.createdAt), 'MM/d/yyyy') : 'N/A'}
            </Description>
            <Description loading={loading} term="Description">
              {repo?.description ? repo.description : 'N/A'}
            </Description>
            <Description term="Size">
              {!loading ? repo?.sizeDisplay : <SkeletonDisplayText />}
            </Description>
          </dl>
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.COMMITS} className={styles.commitsTab}>
          {viewState.globalIdFilter ? (
            !dagsLoading && (
              <CommitDetails commitId={viewState.globalIdFilter} />
            )
          ) : (
            <CommitBrowser repo={repo} repoBaseRef={repoBaseRef} />
          )}
        </Tabs.TabPanel>
      </Tabs.RouterTabs>
    </div>
  );
};

export default RepoDetails;
