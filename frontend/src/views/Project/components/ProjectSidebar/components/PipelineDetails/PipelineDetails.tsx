import {format, fromUnixTime} from 'date-fns';
import React from 'react';
import {Helmet} from 'react-helmet';

import Description from '@dash-frontend/components/Description';
import InfoPanel from '@dash-frontend/components/InfoPanel';
import {SkeletonDisplayText, Tabs} from '@pachyderm/components';

import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails: React.FC = () => {
  const {loading, pipeline, lastJob, isServiceOrSpout, tabsBasePath} =
    usePipelineDetails();

  return (
    <>
      <Helmet>
        <title>Pipeline - Pachyderm Console</title>
      </Helmet>
      <div
        className={styles.sidebarContent}
        data-testid="PipelineDetails__scrollableContent"
      >
        <div className={styles.title}>
          {loading ? (
            <SkeletonDisplayText data-testid="PipelineDetails__pipelineNameSkeleton" />
          ) : (
            <Title>{pipeline?.name}</Title>
          )}
          {pipeline?.description && (
            <div className={styles.description}>{pipeline?.description}</div>
          )}
          <Description term="Most Recent Job Start" loading={loading}>
            {lastJob?.createdAt
              ? format(fromUnixTime(lastJob?.createdAt), 'MMM d, yyyy; h:mma')
              : 'N/A'}
          </Description>
          <Description term="Most Recent Job ID" loading={loading}>
            {lastJob?.id}
          </Description>
        </div>
        <Tabs.RouterTabs basePathTabId={TAB_ID.JOB} basePath={tabsBasePath}>
          <Tabs.TabsHeader className={styles.tabsHeader}>
            {!loading && !isServiceOrSpout && (
              <Tabs.Tab id={TAB_ID.JOB} key={TAB_ID.JOB}>
                Job Overview
              </Tabs.Tab>
            )}
            <Tabs.Tab id={TAB_ID.INFO} key={TAB_ID.INFO}>
              Pipeline Info
            </Tabs.Tab>
            <Tabs.Tab id={TAB_ID.SPEC} key={TAB_ID.SPEC}>
              Spec
            </Tabs.Tab>
          </Tabs.TabsHeader>

          <Tabs.TabPanel id={TAB_ID.JOB}>
            <InfoPanel lastPipelineJob={lastJob} pipelineLoading={loading} />
          </Tabs.TabPanel>
          <Tabs.TabPanel id={TAB_ID.INFO}>
            <PipelineInfo />
          </Tabs.TabPanel>
          <Tabs.TabPanel id={TAB_ID.SPEC}>
            <PipelineSpec />
          </Tabs.TabPanel>
        </Tabs.RouterTabs>
      </div>
    </>
  );
};

export default PipelineDetails;
