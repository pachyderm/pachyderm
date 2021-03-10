import classnames from 'classnames';
import React, {FunctionComponent} from 'react';

import styles from './Card.module.css';
import CardHeader from './components/CardHeader';
import CardTitle from './components/CardTitle';

export type CardProps = React.HTMLAttributes<HTMLDivElement> & {
  autoHeight?: boolean;
  title?: string;
};

export const Card: FunctionComponent<CardProps> = ({
  autoHeight,
  children,
  className,
  title,
  ...props
}) => {
  return (
    <div
      className={classnames(styles.base, className, {
        [styles.autoHeight]: autoHeight,
      })}
      {...props}
    >
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}

      {children}
    </div>
  );
};

export default Object.assign(Card, {Title: CardTitle, Header: CardHeader});
