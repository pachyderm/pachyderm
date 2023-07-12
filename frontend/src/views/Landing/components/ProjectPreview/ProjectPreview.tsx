import {Project, ProjectStatus as ProjectStatusEnum} from '@graphqlTypes';
import classNames from 'classnames';
import React, {useRef} from 'react';
import {Link} from 'react-router-dom';

import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import ProjectJobSetList from '@dash-frontend/components/ProjectJobSetList';
import useIntersection from '@dash-frontend/hooks/useIntersection';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {Group} from '@pachyderm/components';
import getListTitle from 'lib/getListTitle';

import ProjectStatus from '../ProjectStatus';

import styles from './ProjectPreview.module.css';

const emptyProjectMessage = 'Create your first repo/pipeline!';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const {projectDetails, loading, error} = useProjectDetails(project.id, 10);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const subtitleRef = useRef<HTMLHeadingElement>(null);
  const isStuck = useIntersection(subtitleRef.current, sidebarRef.current);

  const shouldShowEmptyState =
    projectDetails?.repoCount === 0 && projectDetails?.pipelineCount === 0;

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
                loading={loading}
                error={error}
              >
                {projectDetails?.repoCount}/{projectDetails?.pipelineCount}
              </Description>
              <Description
                term="Total Data Size"
                loading={loading}
                error={error}
              >
                {projectDetails?.sizeDisplay}
              </Description>
              <Description
                term="Pipeline Status"
                loading={loading}
                error={error}
              >
                <div className={styles.inline}>
                  <ProjectStatus status={project.status} />
                  {project.status.toString() ===
                    ProjectStatusEnum.UNHEALTHY && (
                    <Link
                      to={jobsRoute({projectId: project.id})}
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
          {!loading && (
            <h6
              ref={subtitleRef}
              className={classNames(styles.subTitle, {
                [styles.stuck]: isStuck,
              })}
            >
              {getListTitle('Job', projectDetails?.jobSets?.length || 0)}
            </h6>
          )}
          <ProjectJobSetList
            projectId={project.id}
            jobs={projectDetails?.jobSets}
            loading={loading}
            error={error}
            emptyStateTitle={LETS_START_TITLE}
            emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
          />
        </>
      )}
    </div>
  );
};

export default ProjectPreview;
