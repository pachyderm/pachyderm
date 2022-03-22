import {SkeletonDisplayText, Tabs} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';
import {Helmet} from 'react-helmet';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

import InfoPanel from '../JobDetails/components/InfoPanel';
import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineJobs from './components/PipelineJobs';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails = () => {
  const {viewState} = useUrlQueryState();
  const {loading, pipelineName, filteredTabIds, tabsBasePath} =
    usePipelineDetails();

  return (
    <div className={styles.base}>
      <Helmet>
        <title>Pipeline - Pachyderm Console</title>
      </Helmet>
      <div className={styles.title}>
        {loading ? (
          <SkeletonDisplayText data-testid="PipelineDetails__pipelineNameSkeleton" />
        ) : (
          <Title>{pipelineName}</Title>
        )}
      </div>
      <Tabs.RouterTabs
        basePathTabId={viewState.globalIdFilter ? TAB_ID.JOBS : TAB_ID.INFO}
        basePath={tabsBasePath}
      >
        <Tabs.TabsHeader className={styles.tabsHeader}>
          {filteredTabIds.map((tabId) => {
            let text = capitalize(tabId);
            if (viewState.globalIdFilter && text === 'Jobs') {
              text = 'Job';
            }
            return (
              <Tabs.Tab id={tabId} key={tabId}>
                {text}
              </Tabs.Tab>
            );
          })}
        </Tabs.TabsHeader>

        <Tabs.TabPanel id={TAB_ID.INFO}>
          <PipelineInfo />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.SPEC}>
          <PipelineSpec />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.JOBS}>
          {viewState.globalIdFilter ? <InfoPanel /> : <PipelineJobs />}
        </Tabs.TabPanel>
      </Tabs.RouterTabs>
    </div>
  );
};

export default PipelineDetails;
