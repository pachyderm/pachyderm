import classNames from 'classnames';
import filesize from 'filesize';
import React, {useMemo, useRef} from 'react';
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
import {usePipelines} from '@dash-frontend/hooks/usePipelines';
import {useProjectStatus} from '@dash-frontend/hooks/useProjectStatus';
import {useRepos} from '@dash-frontend/hooks/useRepos';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {Group} from '@pachyderm/components';
import getListTitle from 'lib/getListTitle';

import ProjectStatus from '../ProjectStatus';

import ProjectJobSetList from './components/ProjectJobSetList';
import styles from './ProjectPreview.module.css';

const emptyProjectMessage = 'Create your first repo/pipeline!';

type ProjectPreviewProps = {
  project: ProjectInfo;
};

const JOB_SET_LIMIT = 10;

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const projectName = project.project?.name;
  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useRepos(projectName);
  const {
    pipelines,
    loading: pipelinesLoading,
    error: pipelinesError,
  } = usePipelines(projectName);
  const {
    jobs,
    loading: jobsLoading,
    error: jobsError,
  } = useJobSets({projectName}, JOB_SET_LIMIT);
  const {
    projectStatus,
    loading: projectStatusLoading,
    error: projectStatusError,
  } = useProjectStatus(projectName);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const subtitleRef = useRef<HTMLHeadingElement>(null);
  const isStuck = useIntersection(subtitleRef.current, sidebarRef.current);
  const sizeDisplay = useMemo(() => {
    const totalSizeBytes =
      repos?.reduce(
        (sum, repo) =>
          sum +
          Number(repo?.details?.sizeBytes ?? repo?.sizeBytesUpperBound ?? 0),
        0,
      ) || 0;

    return filesize(totalSizeBytes);
  }, [repos]);
  const repoCount = repos?.length || 0;
  const pipelineCount = pipelines?.length || 0;
  const shouldShowEmptyState = repoCount === 0 && pipelineCount === 0;

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
                loading={reposLoading || pipelinesLoading}
                error={reposError || pipelinesError}
              >
                {repoCount}/{pipelineCount}
              </Description>
              <Description
                term="Total Data Size"
                loading={reposLoading}
                error={reposError}
              >
                {sizeDisplay}
              </Description>
              <Description
                term="Pipeline Status"
                loading={projectStatusLoading}
                error={projectStatusError}
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
              <Group.Divider />
            </>
          )}
        </Group>
      </div>
      {!shouldShowEmptyState && (
        <>
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
        </>
      )}
    </div>
  );
};

export default ProjectPreview;
