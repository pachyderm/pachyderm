import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useCallback, useState} from 'react';

import styles from './Switch.module.css';

type SwitchProps = {
  checked?: boolean;
  className?: string;
  'data-testid'?: string;
  defaultChecked?: boolean;
  id?: string;
  name?: string;
  onChange?: (checked: boolean) => void;
};

const Switch: React.FC<SwitchProps> = ({
  checked: controlledChecked,
  className,
  defaultChecked = false,
  id,
  name,
  onChange = noop,
  'data-testid': testId = 'Switch__button',
  ...props
}) => {
  const controlled = controlledChecked !== undefined;
  const initialChecked = controlled ? controlledChecked : defaultChecked;
  const [checked, setChecked] = useState(initialChecked);
  const handleClick = useCallback(() => {
    if (!controlled) {
      onChange(!checked);
      setChecked(!checked);
    } else {
      onChange(checked);
    }
  }, [checked, controlled, onChange]);

  return (
    <button
      aria-checked={checked}
      className={classNames(
        styles.base,
        {[styles.checked]: checked},
        className,
      )}
      data-testid={testId}
      onClick={handleClick}
      role="switch"
      type="button"
      {...props}
    >
      <div className={styles.track} data-testid={`${testId}Track`}>
        <div className={styles.thumb} data-testid={`${testId}Thumb`} />
      </div>
      <input
        aria-checked={checked}
        checked={checked}
        data-testid={`${testId}Input`}
        id={id}
        name={name}
        readOnly
        role="switch"
        type="hidden"
      />
    </button>
  );
};

export default Switch;
