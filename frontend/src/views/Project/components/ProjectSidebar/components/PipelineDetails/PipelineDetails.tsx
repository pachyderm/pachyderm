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

type PipelineDetailsProps = {
  dagLinks?: Record<string, string[]>;
  dagsLoading?: boolean;
};

const PipelineDetails: React.FC<PipelineDetailsProps> = ({dagLinks}) => {
  const {viewState} = useUrlQueryState();
  const {loading, pipelineName, filteredTabIds, tabsBasePath} =
    usePipelineDetails();

  return (
    <>
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
          <PipelineInfo dagLinks={dagLinks} />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.SPEC}>
          <PipelineSpec />
        </Tabs.TabPanel>
        <Tabs.TabPanel id={TAB_ID.JOBS}>
          {viewState.globalIdFilter ? <InfoPanel /> : <PipelineJobs />}
        </Tabs.TabPanel>
      </Tabs.RouterTabs>
    </>
  );
};

export default PipelineDetails;
