import React from 'react';

import {CodePreview} from 'CodePreview';

import Terminal from '../Terminal';

import Module from './components/Module';
import styles from './ConfigurationUploadModule.module.css';

type File = {
  path: string;
  name: string;
  contents: string;
};

export type ConfigurationUploadConfig = {
  file: File;
};

const ConfigurationUploadModule: React.FC<ConfigurationUploadConfig> = ({
  file,
}) => {
  return (
    <Module title={file.name}>
      <div className={styles.configFile}>
        {file.contents ? <CodePreview>{file.contents}</CodePreview> : null}
      </div>
      <Terminal>{`pachctl create pipeline -f ${file.path}`}</Terminal>
    </Module>
  );
};

export default ConfigurationUploadModule;
