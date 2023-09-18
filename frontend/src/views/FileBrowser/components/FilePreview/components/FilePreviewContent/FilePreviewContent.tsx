import {File} from '@graphqlTypes';
import React from 'react';

import CodePreview, {
  useFileDetails,
} from '@dash-frontend/components/CodePreview';
import EmptyState from '@dash-frontend/components/EmptyState';
import {Link, Switch} from '@pachyderm/components';

import CSVPreview from '../CSVPreview';
import IFramePreview from '../IFramePreview';
import MarkdownPreview from '../MarkdownPreview';
import WebPreview from '../WebPreview';

import styles from './FilePreviewContent.module.css';

type FilePreviewContentProps = {
  download: File['download'];
  path: File['path'];
  viewSource?: boolean;
  toggleViewSource?: () => void;
};

type FilePreviewRendererProps = Omit<
  FilePreviewContentProps,
  'toggleViewSource'
>;

const FilePreviewRenderer = ({
  download,
  path,
  viewSource = false,
}: FilePreviewRendererProps) => {
  const {fileDetails, parsedFilePath} = useFileDetails(path);
  const fileName = parsedFilePath.base;

  if (!download) {
    // TODO: Add a download link when we don't support file previews
    return (
      <EmptyState
        error
        title="Unable to preview this file"
        message="This file is too large to preview"
      />
    );
  }

  if (!fileDetails.supportsPreview) {
    return (
      <>
        <EmptyState
          error
          title="Unable to preview this file"
          message="This file format is not supported for file previews"
        />
        {download && (
          <Link className={styles.viewRaw} href={download} target="_blank">
            View Raw
          </Link>
        )}
      </>
    );
  }

  if (viewSource) {
    return (
      <CodePreview downloadLink={download} language={fileDetails.language} />
    );
  }

  switch (fileDetails.renderer) {
    case 'markdown':
      return <MarkdownPreview downloadLink={download} />;
    case 'image':
      return (
        <img
          className={styles.media}
          src={download}
          alt={fileName}
          data-testid="FilePreviewContent__image"
        />
      );
    case 'video':
      return (
        <video
          className={styles.media}
          controls
          src={download}
          data-testid="FilePreviewContent__video"
        />
      );
    case 'audio':
      return (
        <audio className={styles.media} controls>
          <source src={download} data-testid="FilePreviewContent__audio" />
        </audio>
      );
    case 'web':
      return (
        <WebPreview
          downloadLink={download}
          fileName={fileName}
          data-testid="FilePreviewContent__html"
        />
      );
    case 'iframe':
      return (
        <IFramePreview
          downloadLink={download}
          fileName={fileName}
          data-testid="FilePreviewContent__xml"
        />
      );
    case 'csv':
      return <CSVPreview downloadLink={download} />;
    case 'code':
      return (
        <CodePreview downloadLink={download} language={fileDetails.language} />
      );
    default:
      return null;
  }
};

const FilePreviewContent = ({
  download,
  path,
  viewSource = false,
  toggleViewSource,
}: FilePreviewContentProps) => {
  const {fileDetails} = useFileDetails(path);

  return (
    <>
      <div className={styles.actionBar}>
        {fileDetails.supportsViewSource ? (
          <div className={styles.actionBarItem}>
            <Switch onChange={toggleViewSource} />
            <span className={styles.actionBarViewSourceText}>View Source</span>
          </div>
        ) : null}
      </div>
      <FilePreviewRenderer
        download={download}
        path={path}
        viewSource={viewSource}
      />
    </>
  );
};

export default FilePreviewContent;
