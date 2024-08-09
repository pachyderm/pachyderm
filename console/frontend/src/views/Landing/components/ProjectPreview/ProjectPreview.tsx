import classNames from 'classnames';
import filesize from 'filesize';
import React, {useRef} from 'react';
import {Link} from 'react-router-dom';

import {
  ProjectStatus as ProjectStatusEnum,
  ProjectInfo,
} from '@dash-frontend/api/pfs';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import useIntersection from '@dash-frontend/hooks/useIntersection';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import {usePipelineSummaries} from '@dash-frontend/hooks/usePipelineSummary';
import {useRepoSummary} from '@dash-frontend/hooks/useRepoSummary';
import UserMetadata from '@dash-frontend/views/Project/components/ProjectSidebar/components/RepoDetails/components/UserMetadata';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {Group, Tabs} from '@pachyderm/components';
import getListTitle from 'lib/getListTitle';

import ProjectStatus from '../ProjectStatus';

import ProjectJobSetList from './components/ProjectJobSetList';
import styles from './ProjectPreview.module.css';

export enum TAB_ID {
  OVERVIEW = 'overview',
  METADATA = 'metadata',
}

const emptyProjectMessage = 'Create your first repo/pipeline!';

type ProjectPreviewProps = {
  project: ProjectInfo;
  allProjectNames: (string | undefined)[];
};

const JOB_SET_LIMIT = 10;

const ProjectPreview: React.FC<ProjectPreviewProps> = ({
  project,
  allProjectNames,
}) => {
  const projectName = project.project?.name;
  const {
    repoSummary,
    loading: repoSummaryLoading,
    error: repoSummaryError,
  } = useRepoSummary(projectName);
  const {
    pipelineSummaries,
    loading: pipelineSummaryLoading,
    error: pipelineSummaryError,
  } = usePipelineSummaries(allProjectNames);
  const pipelineSummary = pipelineSummaries?.find(
    (summary) => summary.summary.project?.name === project.project?.name,
  );
  const pipelineCount = pipelineSummary?.pipelineCount;
  const projectStatus = pipelineSummary?.projectStatus;

  const {
    jobs,
    loading: jobsLoading,
    error: jobsError,
  } = useJobSets({projectName}, JOB_SET_LIMIT);

  const sidebarRef = useRef<HTMLDivElement>(null);
  const subtitleRef = useRef<HTMLHeadingElement>(null);
  const isStuck = useIntersection(subtitleRef.current, sidebarRef.current);

  const shouldShowEmptyState =
    Number(repoSummary?.userRepoCount) === 0 && pipelineCount === 0;

  return (
    <div className={styles.base} ref={sidebarRef}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h5>Project Preview</h5>
          {shouldShowEmptyState ? (
            <EmptyState
              title={LETS_START_TITLE}
              message={emptyProjectMessage}
            />
          ) : (
            <>
              <Description
                term="Total No. of Repos/Pipelines"
                loading={repoSummaryLoading || pipelineSummaryLoading}
                error={repoSummaryError || pipelineSummaryError}
              >
                {repoSummary?.userRepoCount || 0}/{pipelineCount}
              </Description>
              <Description
                term="Total Data Size"
                loading={repoSummaryLoading}
                error={repoSummaryError}
              >
                {filesize(Number(repoSummary?.sizeBytes) || 0)}
              </Description>
              <Description
                term="Pipeline Status"
                loading={pipelineSummaryLoading}
                error={pipelineSummaryError}
              >
                <div className={styles.inline}>
                  <ProjectStatus status={projectStatus} />
                  {projectStatus?.toString() ===
                    ProjectStatusEnum.UNHEALTHY && (
                    <Link
                      to={jobsRoute({projectId: project?.project?.name || ''})}
                      className={styles.link}
                    >
                      Inspect
                    </Link>
                  )}
                </div>
              </Description>
            </>
          )}
        </Group>
      </div>
      {!shouldShowEmptyState && (
        <Tabs initialActiveTabId={TAB_ID.OVERVIEW}>
          <Tabs.TabsHeader>
            <div className={styles.tabs}>
              <Tabs.Tab id={TAB_ID.OVERVIEW}>Overview</Tabs.Tab>
              <Tabs.Tab id={TAB_ID.METADATA}>User Metadata</Tabs.Tab>
            </div>
          </Tabs.TabsHeader>
          <Tabs.TabPanel id={TAB_ID.OVERVIEW}>
            {!jobsLoading && (
              <h6
                ref={subtitleRef}
                className={classNames(styles.subTitle, {
                  [styles.stuck]: isStuck,
                })}
              >
                {getListTitle('Job', jobs?.length || 0)}
              </h6>
            )}
            <ProjectJobSetList
              projectId={project?.project?.name || ''}
              jobs={jobs}
              loading={jobsLoading}
              error={jobsError}
              emptyStateTitle={LETS_START_TITLE}
              emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
            />
          </Tabs.TabPanel>
          <Tabs.TabPanel id={TAB_ID.METADATA}>
            <section className={styles.section}>
              <UserMetadata
                metadataType="project"
                id={projectName}
                metadata={project?.metadata}
                editable={true}
              />
            </section>
          </Tabs.TabPanel>
        </Tabs>
      )}
    </div>
  );
};

export default ProjectPreview;
