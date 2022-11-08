import React from 'react';

import {DefaultDropdown, Group, Icon, SettingsSVG} from '@pachyderm/components';

import CodeElement from './components/CodeElement';
import styles from './ConfigFilePreview.module.css';
import useConfigFilePreview from './useConfigFilePreview';
import {Format} from './utils/stringifyToFormat';

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
      <div className={styles.codeBody}>
        <CodeElement element={config} format={Format.YAML} />
      </div>
    </div>
  );
};

export default ConfigPreview;
