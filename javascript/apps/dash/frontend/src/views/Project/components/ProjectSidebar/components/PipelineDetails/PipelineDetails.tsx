import {SkeletonDisplayText, Tabs} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';

import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineJobs from './components/PipelineJobs';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID, TAB_IDS} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails = () => {
  const {
    initialActiveTabId,
    loading,
    pipelineName,
    handleSwitch,
  } = usePipelineDetails();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonDisplayText
          data-testid={'PipelineDetails__pipelineNameSkeleton'}
        />
      ) : (
        <Title>{pipelineName}</Title>
      )}
      <Tabs initialActiveTabId={initialActiveTabId} onSwitch={handleSwitch}>
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
      </Tabs>
    </div>
  );
};

export default PipelineDetails;
