import classNames from 'classnames';
import React, {SVGAttributes} from 'react';

import {ProgressCheckSVG} from '../Svg';

import styles from './SuccessCheckmark.module.css';

interface SuccessCheckmarkProps extends SVGAttributes<HTMLOrSVGElement> {
  show: boolean;
}

const SuccessCheckmark: React.FC<SuccessCheckmarkProps> = ({
  show,
  className,
  ...rest
}) => {
  return (
    (show && (
      <ProgressCheckSVG
        {...rest}
        className={classNames(styles.base, className)}
      />
    )) ||
    null
  );
};

export default SuccessCheckmark;
