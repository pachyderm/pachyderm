import React from 'react';

import useCommit from '@dash-frontend/hooks/useCommit';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Button, RepoSVG} from '@pachyderm/components';

const InspectCommitsButton: React.FC = () => {
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {searchParams} = useUrlQueryState();
  const {repoId, projectId} = useUrlState();

  const {commit} = useCommit({
    args: {
      projectId,
      repoName: repoId,
      id: searchParams.globalIdFilter || '',
    },
  });

  return (
    <Button
      buttonType="secondary"
      IconSVG={RepoSVG}
      to={
        commit
          ? getPathToFileBrowser({
              projectId,
              repoId: repoId,
              commitId: commit.id,
              branchId: commit.branch?.name || 'default',
            })
          : ''
      }
      disabled={!commit}
    >
      Inspect Commits
    </Button>
  );
};

export default InspectCommitsButton;
