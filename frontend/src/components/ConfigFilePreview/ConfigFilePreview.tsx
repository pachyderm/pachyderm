import React from 'react';

import CodePreview from '@dash-frontend/components/CodePreview';
import {DefaultDropdown, Group, Icon, SettingsSVG} from '@pachyderm/components';

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
}

const ConfigPreview: React.FC<ConfigFilePreviewProps> = ({
  config,
  title,
  header = null,
  ...rest
}) => {
  const {onSelectGearAction} = useConfigFilePreview({
    config,
  });

  return (
    <div {...rest}>
      <div className={styles.header}>
        <h5>{title}</h5>
        <Group className={styles.settings} spacing={16}>
          {header}
          <DefaultDropdown
            items={gearDropdownItems}
            menuOpts={{pin: 'right'}}
            buttonOpts={{color: 'purple', hideChevron: true}}
            onSelect={onSelectGearAction}
          >
            <Icon color="plum">
              <SettingsSVG />
            </Icon>
          </DefaultDropdown>
        </Group>
      </div>
      <div
        className={styles.codeBody}
        data-testid="ConfigFilePreview__codeElement"
      >
        <CodePreview
          className={styles.fileCodePreview}
          language="json"
          source={JSON.stringify(config, null, 2)}
          hideLineNumbers
          fullHeight
        />
      </div>
    </div>
  );
};

export default ConfigPreview;
