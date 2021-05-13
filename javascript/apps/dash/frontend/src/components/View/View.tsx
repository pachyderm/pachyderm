import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './View.module.css';

interface ViewProps extends HTMLAttributes<HTMLDivElement> {
  sidenav?: boolean;
  canvas?: boolean;
}

const View: React.FC<ViewProps> = ({
  children,
  className,
  sidenav,
  canvas,
  ...rest
}) => {
  return (
    <div
      className={classnames(
        styles.base,
        {[styles.sidenav]: sidenav},
        {[styles.canvas]: canvas},
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  );
};

export default View;
