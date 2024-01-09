import React from 'react';

import CodePreview, {
  EffectiveSpecPreview,
} from '@dash-frontend/components/CodePreview';
import {
  DefaultDropdown,
  ButtonGroup,
  Button,
  ChevronUpSVG,
  ChevronDownSVG,
  Icon,
  SettingsSVG,
  Tooltip,
} from '@pachyderm/components';

import styles from './ConfigFilePreview.module.css';
import useConfigFilePreview from './useConfigFilePreview';

const gearDropdownItems = [
  {
    id: 'copy',
    content: 'Copy',
    closeOnClick: true,
  },
  {
    id: 'download-json',
    content: 'Download JSON',
    closeOnClick: true,
  },
  {
    id: 'download-yaml',
    content: 'Download YAML',
    closeOnClick: true,
  },
];

interface ConfigFilePreviewProps {
  config: Record<string, unknown>;
  title?: string;
  header?: JSX.Element;
  allowMinimize?: boolean;
  userSpecJSON?: JSON;
}

const ConfigPreview: React.FC<ConfigFilePreviewProps> = ({
  config,
  title,
  header = null,
  allowMinimize = false,
  userSpecJSON,
  ...rest
}) => {
  const {onSelectGearAction, hidden, setHidden} = useConfigFilePreview({
    config,
  });

  return (
    <div {...rest}>
      <div className={styles.header}>
        <h6>{title}</h6>
        <ButtonGroup className={styles.settings}>
          {header}
          {allowMinimize && (
            <Tooltip tooltipText={hidden ? 'Maximize Spec' : 'Minimize Spec'}>
              <Button
                IconSVG={hidden ? ChevronDownSVG : ChevronUpSVG}
                onClick={() => setHidden(!hidden)}
                buttonType="ghost"
                aria-label={hidden ? 'Maximize Spec' : 'Minimize Spec'}
              />
            </Tooltip>
          )}
          <DefaultDropdown
            items={gearDropdownItems}
            menuOpts={{pin: 'right'}}
            buttonOpts={{color: 'purple', hideChevron: true}}
            onSelect={onSelectGearAction}
            aria-label="Pipeline Spec Options"
          >
            <Icon color="plum" small>
              <SettingsSVG />
            </Icon>
          </DefaultDropdown>
        </ButtonGroup>
      </div>
      {(!allowMinimize || !hidden) && (
        <div
          className={styles.codeBody}
          data-testid="ConfigFilePreview__codeElement"
        >
          {userSpecJSON ? (
            <EffectiveSpecPreview
              language="json"
              userSpecJSON={userSpecJSON}
              className={styles.fileCodePreview}
              source={JSON.stringify(config, null, 2)}
              hideLineNumbers
              fullHeight
            />
          ) : (
            <CodePreview
              className={styles.fileCodePreview}
              language="json"
              source={JSON.stringify(config, null, 2)}
              hideLineNumbers
              fullHeight
            />
          )}
        </div>
      )}
    </div>
  );
};

export default ConfigPreview;
