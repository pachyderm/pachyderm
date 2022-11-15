import classNames from 'classnames';
import React, {SVGAttributes} from 'react';

import {CheckmarkSVG} from '../Svg';

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
      <CheckmarkSVG {...rest} className={classNames(styles.base, className)} />
    )) ||
    null
  );
};

export default SuccessCheckmark;
