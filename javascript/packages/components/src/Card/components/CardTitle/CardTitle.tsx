import React, {HTMLAttributes} from 'react';

import styles from './CardTitle.module.css';

export interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement> {
  elemType?: React.ElementType;
}

const CardTitle: React.FC<CardTitleProps> = ({
  className,
  elemType: ElemType = 'h2',
  children,
  ...rest
}) => {
  return (
    <div className={className}>
      <ElemType className={styles.base} {...rest}>
        {children}
      </ElemType>
    </div>
  );
};

export default CardTitle;
