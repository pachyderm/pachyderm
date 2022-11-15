import React, {useCallback} from 'react';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  Link,
  RepoSVG,
  PipelineSVG,
  Icon,
  useClipboardCopy,
  SuccessCheckmark,
  CopySVG,
} from '@pachyderm/components';

import styles from './ResourceLink.module.css';

type ResourceLinkProps = {
  name: string;
};

interface CommitLinkProps extends ResourceLinkProps {
  branch: string;
  repoName: string;
}

export const RepoLink: React.FC<ResourceLinkProps> = ({name, ...rest}) => {
  const {projectId} = useUrlState();
  const nodeName = name.replace(/_repo$/, '');
  const path = repoRoute({
    branchId: 'default',
    projectId,
    repoId: nodeName,
  });

  return (
    <Link
      data-testid="ResourceLink__repo"
      to={path}
      aria-label={nodeName}
      className={styles.base}
    >
      <Icon small className={styles.nodeImage}>
        <RepoSVG />
      </Icon>
      <span {...rest} className={styles.nodeName}>
        {nodeName}
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
  branch,
  repoName,
  ...rest
}) => {
  const {projectId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const path = getPathToFileBrowser({
    projectId,
    commitId: name,
    branchId: branch,
    repoId: repoName,
  });

  return (
    <Link
      data-testid="ResourceLink__commit"
      to={path}
      aria-label={name}
      className={styles.base}
    >
      <span {...rest} className={styles.nodeName}>
        {name}
      </span>
    </Link>
  );
};

export const EgressLink: React.FC<ResourceLinkProps> = ({name, ...rest}) => {
  const {copy, copied, reset} = useClipboardCopy(name);
  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 2000);
  }, [copy, reset]);

  return (
    <button
      data-testid="ResourceLink__egress"
      onClick={handleCopy}
      aria-label={name}
      className={styles.base}
    >
      <Icon small className={styles.nodeImage}>
        <PipelineSVG />
      </Icon>
      <span {...rest} className={styles.nodeName}>
        {name}
      </span>
      <Icon small className={styles.nodeImage}>
        <CopySVG />
      </Icon>
      <Icon small>
        <SuccessCheckmark
          show={copied}
          aria-label={'You have successfully copied the id'}
        />
      </Icon>
    </button>
  );
};
