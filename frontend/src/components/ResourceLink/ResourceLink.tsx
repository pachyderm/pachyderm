import React from 'react';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Link, RepoSVG, PipelineSVG, Icon, IdText} from '@pachyderm/components';

import styles from './ResourceLink.module.css';

type ResourceLinkProps = {
  name: string;
  projectId?: string;
};

interface CommitLinkProps extends ResourceLinkProps {
  repoName: string;
}

export const RepoLink: React.FC<ResourceLinkProps> = ({
  name,
  projectId,
  ...rest
}) => {
  const {projectId: projectIdFromURL} = useUrlState();
  const path = repoRoute({
    projectId: projectId ?? projectIdFromURL,
    repoId: name,
  });

  const repoText =
    projectId && projectId !== projectIdFromURL
      ? `${name} (Project ${projectId})`
      : name;

  return (
    <Link
      data-testid="ResourceLink__repo"
      to={path}
      aria-label={name}
      className={styles.base}
    >
      <Icon small className={styles.nodeImage}>
        <RepoSVG />
      </Icon>
      <span {...rest} className={styles.nodeName}>
        {repoText}
      </span>
    </Link>
  );
};

export const PipelineLink: React.FC<ResourceLinkProps> = ({name, ...rest}) => {
  const {projectId} = useUrlState();
  const path = pipelineRoute({
    projectId,
    pipelineId: name,
  });

  return (
    <Link
      data-testid="ResourceLink__pipeline"
      to={path}
      aria-label={name}
      className={styles.base}
    >
      <Icon small className={styles.nodeImage}>
        <PipelineSVG />
      </Icon>
      <span {...rest} className={styles.nodeName}>
        {name}
      </span>
    </Link>
  );
};

export const CommitLink: React.FC<CommitLinkProps> = ({
  name,
  repoName,
  ...rest
}) => {
  const {projectId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const path = getPathToFileBrowser({
    projectId,
    commitId: name,
    repoId: repoName,
  });

  return (
    <Link
      data-testid="ResourceLink__commit"
      to={path}
      aria-label={name}
      className={styles.base}
    >
      <IdText {...rest} className={styles.nodeName}>
        {name}
      </IdText>
    </Link>
  );
};
