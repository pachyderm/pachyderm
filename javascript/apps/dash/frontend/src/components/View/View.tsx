import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './View.module.css';

interface ViewProps extends HTMLAttributes<HTMLDivElement> {
  sidenav?: boolean;
}

const View: React.FC<ViewProps> = ({children, className, sidenav, ...rest}) => {
  return (
    <div
      className={classnames(
        styles.base,
        {[styles.sidenav]: sidenav},
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  );
};

export default View;
