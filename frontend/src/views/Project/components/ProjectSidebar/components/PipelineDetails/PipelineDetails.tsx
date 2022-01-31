import {
  DocumentSVG,
  Link,
  SkeletonDisplayText,
  Tabs,
  Icon,
} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';

import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineJobs from './components/PipelineJobs';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails = () => {
  const {
    loading,
    pipelineName,
    filteredTabIds,
    pipelineLogsRoute,
    tabsBasePath,
  } = usePipelineDetails();

  return (
    <div className={styles.base}>
      <div className={styles.title}>
        {loading ? (
          <SkeletonDisplayText data-testid="PipelineDetails__pipelineNameSkeleton" />
        ) : (
          <Title>{pipelineName}</Title>
        )}
      </div>
      <Tabs.RouterTabs basePathTabId={TAB_ID.INFO} basePath={tabsBasePath}>
        <Tabs.TabsHeader className={styles.tabsHeader}>
          {filteredTabIds.map((tabId) => (
            <Tabs.Tab id={tabId} key={tabId}>
              {capitalize(tabId)}
            </Tabs.Tab>
          ))}
          <li className={styles.logsWrapper}>
            <Link small to={pipelineLogsRoute}>
              <span className={styles.readLogsText}>
                Read Logs{' '}
                <Icon small color="plum">
                  <DocumentSVG />
                </Icon>
              </span>
            </Link>
          </li>
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
