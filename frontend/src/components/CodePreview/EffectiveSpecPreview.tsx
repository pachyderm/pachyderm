import classnames from 'classnames';
import React from 'react';

import CodePreview, {CodePreviewProps} from './CodePreview';
import styles from './EffectiveSpecPreview.module.css';
import {dynamicEffectiveSpecDecorations} from './extensions/dynamicEffectiveSpecDecorations';
import {createEffectiveSpecDecorationMap} from './utils/createEffectiveSpecDecorationMap';

type EffectiveSpecPreviewProps = {
  userSpecJSON: JSON | undefined;
  clusterDefaultsJSON?: JSON;
};

const EffectiveSpecPreview: React.FC<
  Omit<CodePreviewProps, 'extension'> & EffectiveSpecPreviewProps
> = ({userSpecJSON, clusterDefaultsJSON, className, ...rest}) => {
  const userOverrides = createEffectiveSpecDecorationMap(
    userSpecJSON as unknown as Record<string, unknown>,
    clusterDefaultsJSON as unknown as Record<string, unknown>,
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
