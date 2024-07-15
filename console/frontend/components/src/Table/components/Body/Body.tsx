import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

const Body: React.FC<HTMLAttributes<HTMLTableSectionElement>> = ({
  children,
  className,
  ...rest
}) => (
  <tbody className={classnames(className)} {...rest}>
    {children}
  </tbody>
);

export default Body;
