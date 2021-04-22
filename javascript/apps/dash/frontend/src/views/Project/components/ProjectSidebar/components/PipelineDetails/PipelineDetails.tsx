import {SkeletonDisplayText, Tabs} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';

import {PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineJobs from './components/PipelineJobs';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID, TAB_IDS} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails = () => {
  const {loading, pipelineName} = usePipelineDetails();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonDisplayText
          data-testid={'PipelineDetails__pipelineNameSkeleton'}
        />
      ) : (
        <Title>{pipelineName}</Title>
      )}
      <Tabs.RouterTabs basePathTabId={TAB_ID.INFO} basePath={PIPELINE_PATH}>
        <Tabs.TabsHeader className={styles.tabsHeader}>
          {TAB_IDS.map((tabId) => (
            <Tabs.Tab id={tabId} key={tabId}>
              {capitalize(tabId)}
            </Tabs.Tab>
          ))}
        </Tabs.TabsHeader>

        <Tabs.TabPanel id={TAB_ID.INFO}>
          <PipelineInfo />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.SPEC}>
          <PipelineSpec />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.JOBS}>
          <PipelineJobs />
        </Tabs.TabPanel>
      </Tabs.RouterTabs>
    </div>
  );
};

export default PipelineDetails;
