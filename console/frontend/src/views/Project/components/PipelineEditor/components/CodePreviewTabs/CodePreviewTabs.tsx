import React from 'react';

import CodePreview, {
  EffectiveSpecPreview,
} from '@dash-frontend/components/CodePreview';
import {Tabs} from '@pachyderm/components';

import styles from '../../PipelineEditor.module.css';

type CodePreviewTypes = {
  editorTextJSON?: JSON;
  clusterDefaultsJSON?: JSON;
  projectDefaultsJSON?: JSON;
  effectiveSpec?: string;
  clusterDefaults?: string;
  projectDefaults?: string;
  effectiveSpecLoading: boolean;
  clusterDefaultsLoading: boolean;
  projectDefaultsLoading: boolean;
};

const CodePreviewTabs: React.FC<CodePreviewTypes> = ({
  editorTextJSON,
  clusterDefaultsJSON,
  projectDefaultsJSON,
  effectiveSpec,
  effectiveSpecLoading,
  clusterDefaults,
  clusterDefaultsLoading,
  projectDefaults,
  projectDefaultsLoading,
}) => (
  <>
    <Tabs.TabPanel id="effective-spec" className={styles.editorTabPanel}>
      <EffectiveSpecPreview
        language="json"
        userSpecJSON={editorTextJSON}
        clusterDefaultsJSON={clusterDefaultsJSON}
        projectDefaultsJSON={projectDefaultsJSON}
        source={effectiveSpec}
        sourceLoading={
          (!effectiveSpec && effectiveSpecLoading) || clusterDefaultsLoading
        }
        className={styles.codePreview}
      />
    </Tabs.TabPanel>
    <Tabs.TabPanel id="cluster-defaults" className={styles.editorTabPanel}>
      <CodePreview
        language="json"
        source={clusterDefaults}
        sourceLoading={clusterDefaultsLoading}
        className={styles.codePreview}
      />
    </Tabs.TabPanel>
    <Tabs.TabPanel id="project-defaults" className={styles.editorTabPanel}>
      <CodePreview
        language="json"
        source={projectDefaults}
        sourceLoading={projectDefaultsLoading}
        className={styles.codePreview}
      />
    </Tabs.TabPanel>
  </>
);

export default CodePreviewTabs;
