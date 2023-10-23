import React from 'react';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import InfoPanel from '@dash-frontend/components/InfoPanel';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useCurrentOuptutRepoOfPipeline from '@dash-frontend/hooks/useCurrentOuptutRepoOfPipeline';
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
    pipelineAndJobloading,
    pipeline,
    lastJob,
    isServiceOrSpout,
    isSpout,
    tabsBasePath,
    projectId,
    pipelineId,
    editRolesPermission,
    pipelineReadPermission,
    globalId,
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
          {pipelineAndJobloading ? (
            <SkeletonDisplayText />
          ) : (
            <Title>{pipeline?.name}</Title>
          )}
          {pipeline?.description && (
            <div className={styles.description}>{pipeline?.description}</div>
          )}
          {repo?.authInfo?.rolesList && (
            <Description term="Your Roles" loading={pipelineAndJobloading}>
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
          {!isSpout && pipelineReadPermission && (
            <>
              <Description
                term={globalId ? 'Global ID' : 'Most Recent Job ID'}
                loading={pipelineAndJobloading}
              >
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
        {!pipelineAndJobloading && (
          <Tabs.RouterTabs
            basePathTabId={!isServiceOrSpout ? TAB_ID.JOB : TAB_ID.INFO}
            basePath={tabsBasePath}
          >
            <Tabs.TabsHeader className={styles.tabsHeader}>
              {!isServiceOrSpout && (
                <Tabs.Tab id={TAB_ID.JOB} key={TAB_ID.JOB}>
                  Job Overview
                </Tabs.Tab>
              )}
              <Tabs.Tab id={TAB_ID.INFO} key={TAB_ID.INFO}>
                Pipeline Info
              </Tabs.Tab>
              {!globalId && (
                <Tabs.Tab id={TAB_ID.SPEC} key={TAB_ID.SPEC}>
                  Spec
                </Tabs.Tab>
              )}
            </Tabs.TabsHeader>
            {!isServiceOrSpout && (
              <Tabs.TabPanel id={TAB_ID.JOB}>
                {pipelineReadPermission ? (
                  <InfoPanel className={styles.paddingUnset} />
                ) : (
                  <EmptyState
                    title={
                      <>{`You don't have permission to view this pipeline`}</>
                    }
                    noAccess
                    message={
                      <>
                        {`You'll need a role of repoReader or higher on the connected repo to view job data
              about this pipeline.`}
                      </>
                    }
                    linkToDocs={{
                      text: 'Read more about authorization',
                      pathWithoutDomain: '/set-up/authorization/',
                    }}
                    className={styles.emptyState}
                  />
                )}
              </Tabs.TabPanel>
            )}
            <Tabs.TabPanel id={TAB_ID.INFO}>
              <PipelineInfo />
            </Tabs.TabPanel>
            {!globalId && (
              <Tabs.TabPanel id={TAB_ID.SPEC}>
                <PipelineSpec />
              </Tabs.TabPanel>
            )}
          </Tabs.RouterTabs>
        )}
      </div>
    </>
  );
};

export default PipelineDetails;
