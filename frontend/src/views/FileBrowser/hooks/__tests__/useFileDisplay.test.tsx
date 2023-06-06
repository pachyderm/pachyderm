import {File, FileType} from '@graphqlTypes';
import {screen, render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useFileDisplay from '../useFileDisplay';

describe('FileBrowser/hooks/useFileDisplay', () => {
  const Component = withContextProviders(({file}: {file: File}) => {
    const fileDisplay = useFileDisplay(file);

    return (
      <span>previewSupported: {fileDisplay.previewSupported.toString()}</span>
    );
  });

  let file: File;

  beforeEach(() => {
    file = {
      commitId: 'd350c8d08a644ed5b2ee98c035ab6b34',
      download: '',
      hash: '',
      repoName: 'coolRepo',
      sizeBytes: 777,
      sizeDisplay: '777 B',
      path: '',
      type: FileType.FILE,
    };
    window.history.replaceState(
      {},
      '',
      `/project/cool/repos/${file.repoName}/branch/master/commit/${file.commitId}`,
    );
  });

  it('should show yml files renderable', () => {
    file.path = 'train.yml';
    render(<Component file={file} />);
    expect(screen.getByText('previewSupported: true')).toBeInTheDocument();
  });

  it('should show yaml files renderable', () => {
    file.path = 'train.yaml';
    render(<Component file={file} />);
    expect(screen.getByText('previewSupported: true')).toBeInTheDocument();
  });

  it('should show lmy files not renderable', () => {
    file.path = 'train.lmy';
    render(<Component file={file} />);
    expect(screen.getByText('previewSupported: false')).toBeInTheDocument();
  });
});
