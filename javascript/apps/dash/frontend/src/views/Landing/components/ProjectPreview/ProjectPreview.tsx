import {Group, InfoSVG} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useRef} from 'react';
import {Link} from 'react-router-dom';

import Description from '@dash-frontend/components/Description';
import JobListStatic from '@dash-frontend/components/JobList/components/JobListStatic';
import useIntersection from '@dash-frontend/hooks/useIntersection';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {Project} from '@graphqlTypes';

import ProjectStatus from '../ProjectStatus';

import styles from './ProjectPreview.module.css';

const emptyStateTitle = "Let's Start :)";
const emptyJobListMessage =
  'Create your first job! If there are any pipeline errors, fix those before you create a job.';

type ProjectPreviewProps = {
  project: Project;
};

const ProjectPreview: React.FC<ProjectPreviewProps> = ({project}) => {
  const {projectDetails, loading} = useProjectDetails(project.id);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const subtitleRef = useRef<HTMLHeadingElement>(null);
  const isStuck = useIntersection(subtitleRef.current, sidebarRef.current);

  return (
    <div className={styles.base} ref={sidebarRef}>
      <div className={styles.topContent}>
        <Group spacing={24} vertical>
          <h4 className={styles.title}>Project Preview</h4>
          <Description term="Total No. of Repos/Pipelines" loading={loading}>
            {projectDetails?.repoCount}/{projectDetails?.repoCount}
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
                  {/* TODO: Link to jobs page */}
                  <Link to={'/'} className={styles.link}>
                    Inspect
                  </Link>
                </span>
              )}
            </div>
          </Description>
          <Group.Divider />
        </Group>
      </div>
      <h4
        ref={subtitleRef}
        className={classNames(styles.subTitle, {
          [styles.stuck]: isStuck,
        })}
      >
        Last 30 Jobs
      </h4>
      <JobListStatic
        projectId={project.id}
        pipelineJobs={projectDetails?.jobs}
        loading={loading}
        emptyStateTitle={emptyStateTitle}
        emptyStateMessage={emptyJobListMessage}
      />
    </div>
  );
};

export default ProjectPreview;
