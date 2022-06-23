import {
  Link,
  RepoSVG,
  PipelineSVG,
  CopySVG,
  Icon,
  useClipboardCopy,
  SuccessCheckmark,
} from '@pachyderm/components';
import React, {useMemo, useCallback} from 'react';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './RepoPipelineLink.module.css';

type RepoPipelineLinkProps = {
  name: string;
  branch?: string;
  repo?: string;
  type: 'repo' | 'pipeline' | 'egress' | 'commit';
};

const getIcon = (type: RepoPipelineLinkProps['type']) => {
  switch (type) {
    case 'repo':
      return <RepoSVG />;
    case 'pipeline':
      return <PipelineSVG />;
    case 'egress':
      return <PipelineSVG />;
    default:
      return undefined;
  }
};

const RepoPipelineLink: React.FC<RepoPipelineLinkProps> = ({
  name,
  type,
  branch = 'master',
  repo = '',
  ...rest
}) => {
  const {projectId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {copy, copied, reset} = useClipboardCopy(name);
  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 2000);
  }, [copy, reset]);

  const nodeName = type !== 'pipeline' ? name.replace('_repo', '') : name;
  const path = useMemo(() => {
    switch (type) {
      case 'repo':
        return repoRoute({
          branchId: 'master',
          projectId,
          repoId: nodeName,
        });
      case 'pipeline':
        return pipelineRoute({
          projectId,
          pipelineId: nodeName,
        });
      case 'commit':
        return getPathToFileBrowser({
          projectId,
          commitId: nodeName,
          branchId: branch,
          repoId: repo,
        });

      default:
        return '/not-found';
    }
  }, [nodeName, projectId, type, getPathToFileBrowser, branch, repo]);

  const icon = useMemo(() => getIcon(type), [type]);

  if (type === 'egress') {
    return (
      <button
        data-testid="RepoPipelineLink__item"
        onClick={handleCopy}
        aria-label={nodeName}
        className={styles.base}
      >
        {icon ? (
          <Icon small className={styles.nodeImage}>
            {icon}
          </Icon>
        ) : null}
        <span {...rest} className={styles.nodeName}>
          {nodeName}
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
  }

  return (
    <Link
      data-testid="RepoPipelineLink__item"
      to={path}
      aria-label={nodeName}
      className={styles.base}
    >
      {icon ? (
        <Icon small className={styles.nodeImage}>
          {icon}
        </Icon>
      ) : null}
      <span {...rest} className={styles.nodeName}>
        {nodeName}
      </span>
    </Link>
  );
};

export default RepoPipelineLink;
