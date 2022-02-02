import React from 'react';

import {PureCheckbox} from 'Checkbox';
import {SuccessCheckmark} from 'SuccessCheckmark';

import Terminal from '../Terminal';

import Module from './components/Module';
import styles from './MultiSelectModule.module.css';

type File = {
  path: string;
  name: string;
  selected: boolean;
  uploaded: boolean;
};

export type MultiSelectConfig = Record<string, File>;

type MultiSelectModuleProps = {
  type: 'file' | 'image';
  files: Record<string, File>;
  onChange: (key: string) => void;
  branch?: string;
  repo: string;
  disabled: boolean;
};

const MultiSelectModule: React.FC<MultiSelectModuleProps> = ({
  disabled,
  files,
  onChange,
  type,
  repo,
  branch = 'master',
}) => {
  const isFile = type === 'file';

  return (
    <Module title={isFile ? 'Select Files' : 'Select Images'}>
      {isFile ? (
        <div className={styles.fileList}>
          {Object.keys(files).map((url) => {
            return (
              <div className={styles.fileRow} key={url}>
                {files[url].uploaded ? (
                  <div className={styles.uploadedFile}>
                    <SuccessCheckmark show />
                    {files[url].name}
                  </div>
                ) : (
                  <PureCheckbox
                    disabled={disabled}
                    selected={files[url].selected || files[url].uploaded}
                    name={files[url].name}
                    data-testid="MultiSelectModule__check"
                    label={
                      <span className={styles.fileCheckboxLabel}>
                        {files[url].name}
                      </span>
                    }
                    onChange={() => onChange(url)}
                    className={
                      files[url].selected ? styles.checked : styles.unchecked
                    }
                  />
                )}
              </div>
            );
          })}
        </div>
      ) : (
        <div className={styles.grid}>
          {Object.keys(files).map((url) => {
            return (
              <div className={styles.imageWrapper} key={url}>
                <div className={styles.imageStatus}>
                  {files[url].uploaded ? (
                    <SuccessCheckmark show />
                  ) : (
                    <PureCheckbox
                      disabled={disabled}
                      selected={files[url].selected || files[url].uploaded}
                      name={files[url].name}
                      label={files[url].name}
                      onChange={() => onChange(url)}
                      className={
                        files[url].selected ? styles.checked : styles.unchecked
                      }
                    />
                  )}
                </div>

                <img src={url} className={styles.image} alt={files[url].name} />
              </div>
            );
          })}
        </div>
      )}
      <Terminal>
        {Object.keys(files).map((url) => {
          if (files[url].selected)
            return (
              <span
                key={url}
              >{`pachctl put file ${repo}@${branch}:${files[url].name} -f ${url}`}</span>
            );
          return null;
        })}
      </Terminal>
    </Module>
  );
};

export default MultiSelectModule;
