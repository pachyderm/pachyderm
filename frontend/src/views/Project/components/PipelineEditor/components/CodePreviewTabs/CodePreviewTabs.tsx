import React from 'react';

import CodePreview, {
  EffectiveSpecPreview,
} from '@dash-frontend/components/CodePreview';
import {Tabs} from '@pachyderm/components';

import styles from '../../PipelineEditor.module.css';

type CodePreviewTypes = {
  editorTextJSON?: JSON;
  clusterDefaultsJSON?: JSON;
  effectiveSpec?: string;
  clusterDefaults?: string;
  effectiveSpecLoading: boolean;
  clusterDefaultsLoading: boolean;
};

const CodePreviewTabs: React.FC<CodePreviewTypes> = ({
  editorTextJSON,
  clusterDefaultsJSON,
  effectiveSpec,
  effectiveSpecLoading,
  clusterDefaults,
  clusterDefaultsLoading,
}) => (
  <>
    <Tabs.TabPanel id="effective-spec" className={styles.editorTabPanel}>
      <EffectiveSpecPreview
        language="json"
        userSpecJSON={editorTextJSON}
        clusterDefaultsJSON={clusterDefaultsJSON}
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
  </>
);

export default CodePreviewTabs;
