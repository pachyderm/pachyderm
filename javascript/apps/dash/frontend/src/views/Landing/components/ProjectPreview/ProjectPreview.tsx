import {Group, InfoSVG} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useRef} from 'react';
import {Link} from 'react-router-dom';

import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobListStatic from '@dash-frontend/components/JobList/components/JobListStatic';
import useIntersection from '@dash-frontend/hooks/useIntersection';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {Project} from '@graphqlTypes';
import getListTitle from 'lib/getListTitle';

import ProjectStatus from '../ProjectStatus';

import styles from './ProjectPreview.module.css';

const emptyProjectMessage = 'Create your first repo/pipeline!';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const {projectDetails, loading} = useProjectDetails(project.id);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const subtitleRef = useRef<HTMLHeadingElement>(null);
  const isStuck = useIntersection(subtitleRef.current, sidebarRef.current);

  const shouldShowEmptyState =
    projectDetails?.repoCount === 0 && projectDetails?.pipelineCount === 0;

  return (
    <div className={styles.base} ref={sidebarRef}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h4 className={styles.title}>Project Preview</h4>
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
              >
                {projectDetails?.repoCount}/{projectDetails?.pipelineCount}
              </Description>
              <Description term="Total Data Size" loading={loading}>
                {projectDetails?.sizeDisplay}
              </Description>
              <Description term="Pipeline Status" loading={loading}>
                <div className={styles.inline}>
                  <ProjectStatus status={project.status} />
                  {project.status.toString() === 'UNHEALTHY' && (
                    <span className={styles.extras}>
                      {/* TODO: add tooltip for unhealthy project details */}
                      <InfoSVG />
                      <Link
                        to={jobsRoute({projectId: project.id})}
                        className={styles.link}
                      >
                        Inspect
                      </Link>
                    </span>
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
            <h4
              ref={subtitleRef}
              className={classNames(styles.subTitle, {
                [styles.stuck]: isStuck,
              })}
            >
              {getListTitle('Job', projectDetails?.jobs?.length || 0)}
            </h4>
          )}
          <JobListStatic
            projectId={project.id}
            jobs={projectDetails?.jobs}
            loading={loading}
            emptyStateTitle={LETS_START_TITLE}
            emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
          />
        </>
      )}
    </div>
  );
};

export default ProjectPreview;
