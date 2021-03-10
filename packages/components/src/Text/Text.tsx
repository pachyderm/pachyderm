import classnames from 'classnames';
import React from 'react';

import useSplitting from '../hooks/useSplitting';

import styles from './Text.module.css';

const Heading: React.FC = (props) => {
  return (
    <h3 className={styles.headingText}>
      <span>{props.children}</span>
    </h3>
  );
};

const Body: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <span
      className={`${styles.bodyText} ${className ? className : ''}`}
      {...rest}
    >
      {children}
    </span>
  );
};

const BodyEm: React.FC = (props) => {
  return (
    <Body>
      <em>{props.children}</em>
    </Body>
  );
};

const BodyError: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...rest
}) => {
  return (
    <span
      className={classnames(styles.bodyErrorText, className)}
      role="status"
      {...rest}
    >
      {children}
    </span>
  );
};

const FieldError: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  children,
  ...rest
}) => (
  <BodyError className={styles.fieldError} role="alert" {...rest}>
    {children}
  </BodyError>
);

const BodyMono: React.FC = (props) => {
  return <code className={styles.bodyMonoText}>{props.children}</code>;
};

const BodyMonoBlock: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  ...rest
}) => {
  return (
    <div className={styles.bodyMonoBlock} {...rest}>
      <BodyMono>{children}</BodyMono>
    </div>
  );
};

const BodyStrong: React.FC = (props) => {
  return (
    <Body>
      <strong>{props.children}</strong>
    </Body>
  );
};

const SlideIn: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  children,
  className,
  ...props
}) => {
  const splitting = useSplitting();
  const classes = classnames(styles.slideIn, className);

  return (
    <span className={classes} {...splitting} {...props}>
      {children}
    </span>
  );
};

export {
  Heading,
  Body,
  BodyEm,
  BodyError,
  BodyMono,
  BodyMonoBlock,
  BodyStrong,
  FieldError,
  SlideIn,
};
