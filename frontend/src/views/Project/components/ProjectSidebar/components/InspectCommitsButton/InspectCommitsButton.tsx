import React from 'react';

import {useCommits} from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Button, NavigationHistorySVG} from '@pachyderm/components';

const InspectCommitsButton: React.FC = () => {
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {searchParams} = useUrlQueryState();
  const {repoId, projectId} = useUrlState();

  const {commits} = useCommits({
    projectName: projectId,
    repoName: repoId,
    args: {
      commitIdCursor: searchParams.globalIdFilter || '',
      number: 1,
    },
  });

  const commit = commits && commits[0];

  return (
    <Button
      buttonType="primary"
      IconSVG={NavigationHistorySVG}
      to={
        commit
          ? getPathToFileBrowser({
              projectId,
              repoId: repoId,
              commitId: commit.commit?.id || '',
              branchId: commit.commit?.branch?.name || 'default',
            })
          : ''
      }
      disabled={!commit}
    >
      Previous Commits
    </Button>
  );
};

export default InspectCommitsButton;
