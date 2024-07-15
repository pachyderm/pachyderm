import classnames from 'classnames';
import merge from 'lodash/merge';
import React from 'react';

import CodePreview, {CodePreviewProps} from './CodePreview';
import styles from './EffectiveSpecPreview.module.css';
import {dynamicEffectiveSpecDecorations} from './extensions/dynamicEffectiveSpecDecorations';
import {createEffectiveSpecDecorationMap} from './utils/createEffectiveSpecDecorationMap';

type EffectiveSpecPreviewProps = {
  userSpecJSON: JSON | undefined;
  clusterDefaultsJSON?: JSON;
  projectDefaultsJSON?: JSON;
};

const EffectiveSpecPreview: React.FC<
  Omit<CodePreviewProps, 'extension'> & EffectiveSpecPreviewProps
> = ({
  userSpecJSON,
  clusterDefaultsJSON,
  projectDefaultsJSON,
  className,
  ...rest
}) => {
  const mergedPipelineDefaults = merge(
    clusterDefaultsJSON,
    projectDefaultsJSON,
  );

  const userOverrides = createEffectiveSpecDecorationMap(
    userSpecJSON as unknown as Record<string, unknown>,
    mergedPipelineDefaults as unknown as Record<string, unknown>,
  );

  return (
    <CodePreview
      className={classnames(className, styles.codePadding)}
      additionalExtensions={[dynamicEffectiveSpecDecorations(userOverrides)]}
      {...rest}
    />
  );
};

export default EffectiveSpecPreview;
