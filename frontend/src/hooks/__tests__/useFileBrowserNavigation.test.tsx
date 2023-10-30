import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import useFileBrowserNavigation from '../useFileBrowserNavigation';

const NavigationToComponent = withContextProviders(
  ({args}: {args: Parameters<typeof fileBrowserRoute>[0]}) => {
    const {getPathToFileBrowser} = useFileBrowserNavigation();
    const pathTo = getPathToFileBrowser(args);

    return (
      <div>
        <span>path to file browser: {pathTo}</span>
      </div>
    );
  },
);

const NavigationFromComponent = withContextProviders(
  ({backupPath}: {backupPath: string}) => {
    const {getPathFromFileBrowser} = useFileBrowserNavigation();
    const pathFrom = getPathFromFileBrowser(backupPath);

    return (
      <div>
        <span>path from file browser: {pathFrom}</span>
      </div>
    );
  },
);

describe('useFileBrowserNavigation', () => {
  it('should get the correct path to the file browser', async () => {
    const backPath = '/project/projectA/repos/repoA';
    const args = {
      projectId: 'projectA',
      repoId: 'repoA',
      commitId: '1234',
      branchId: 'master',
      filePath: '/cats',
    };
    window.history.pushState('', '', backPath);

    render(<NavigationToComponent args={args} />);
    expect(
      screen.getByText(
        `path to file browser: /project/${args.projectId}/repos/${
          args.repoId
        }/branch/${args.branchId}/commit/${args.commitId}/${encodeURIComponent(
          args.filePath,
        )}/?prevPath=${encodeURIComponent(backPath)}`,
      ),
    ).toBeInTheDocument();
  });

  it('should get the correct path from the file browser', async () => {
    const backPath = '/backupPath';
    const path =
      '/project/projectA/repos/repoA/branch/master/commit/1234/%2Fcats/?prevPath=%2Fproject%2FprojectA%2Frepos%2FrepoA';
    window.history.pushState('', '', path);

    render(<NavigationFromComponent backupPath={backPath} />);
    expect(
      screen.getByText('path from file browser: /project/projectA/repos/repoA'),
    ).toBeInTheDocument();
  });

  it('should get the correct backup path from the file browser if prevPath is not available', async () => {
    const backPath = '/backupPath';
    const path =
      '/project/projectA/repos/repoA/branch/master/commit/1234/%2Fcats/';
    window.history.pushState('', '', path);

    render(<NavigationFromComponent backupPath={backPath} />);
    expect(
      screen.getByText('path from file browser: /backupPath'),
    ).toBeInTheDocument();
  });
});
