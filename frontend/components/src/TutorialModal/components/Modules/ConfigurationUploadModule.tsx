import React from 'react';

import {CodePreview} from '@pachyderm/components';

import Terminal from '../Terminal';

import Module from './components/Module';
import styles from './ConfigurationUploadModule.module.css';

type FileMeta = {
  path: string;
  name: string;
};

export type ConfigurationUploadConfig = {
  children?: React.ReactNode;
  fileMeta: FileMeta;
  fileContents?: string;
};

const ConfigurationUploadModule: React.FC<ConfigurationUploadConfig> = ({
  fileMeta,
  fileContents,
  children,
}) => {
  return (
    <Module title={fileMeta.name}>
      <div className={styles.configFile}>
        {fileContents ? <CodePreview>{fileContents}</CodePreview> : children}
      </div>
      <Terminal>{`pachctl create pipeline -f ${fileMeta.path}`}</Terminal>
    </Module>
  );
};

export default ConfigurationUploadModule;
