import React from 'react';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import Description from '@dash-frontend/components/Description';
import InfoPanel from '@dash-frontend/components/InfoPanel';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useCurrentOuptutRepoOfPipeline from '@dash-frontend/hooks/useCurrentOuptutRepoOfPipeline';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  SkeletonDisplayText,
  Tabs,
  Group,
  ButtonLink,
  useModal,
} from '@pachyderm/components';

import Title from '../Title';

import PipelineInfo from './components/PipelineInfo';
import PipelineSpec from './components/PipelineSpec';
import {TAB_ID} from './constants/tabIds';
import usePipelineDetails from './hooks/usePipelineDetails';
import styles from './PipelineDetails.module.css';

const PipelineDetails: React.FC = () => {
  const {
    loading,
    pipeline,
    lastJob,
    isServiceOrSpout,
    isSpout,
    tabsBasePath,
    projectId,
    pipelineId,
    editRolesPermission,
  } = usePipelineDetails();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const {repo} = useCurrentOuptutRepoOfPipeline();
  return (
    <>
      <BrandedTitle title="Pipeline" />
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
          {repo?.authInfo?.rolesList && (
            <Description term="Your Roles" loading={loading}>
              <Group spacing={8}>
                {repo?.authInfo?.rolesList.join(', ') || 'None'}
                <ButtonLink onClick={openRolesModal}>
                  {editRolesPermission
                    ? 'Set Roles via Repo'
                    : 'See All Roles via Repo'}
                </ButtonLink>{' '}
              </Group>
            </Description>
          )}
          {!isSpout && (
            <>
              <Description term="Most Recent Job Start" loading={loading}>
                {lastJob?.createdAt
                  ? getStandardDate(lastJob?.createdAt)
                  : 'N/A'}
              </Description>
              <Description term="Most Recent Job ID" loading={loading}>
                {lastJob?.id}
              </Description>
            </>
          )}
          {repo?.authInfo?.rolesList && rolesModalOpen && (
            <RepoRolesModal
              show={rolesModalOpen}
              onHide={closeRolesModal}
              projectName={projectId}
              repoName={pipelineId}
              readOnly={!editRolesPermission}
            />
          )}
        </div>
        <Tabs.RouterTabs
          basePathTabId={!isServiceOrSpout ? TAB_ID.JOB : TAB_ID.INFO}
          basePath={tabsBasePath}
        >
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
          {!loading && !isServiceOrSpout && (
            <Tabs.TabPanel id={TAB_ID.JOB}>
              <InfoPanel lastPipelineJob={lastJob} pipelineLoading={loading} />
            </Tabs.TabPanel>
          )}
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
