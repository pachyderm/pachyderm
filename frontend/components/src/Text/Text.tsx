import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import useSplitting from '../hooks/useSplitting';

import styles from './Text.module.css';

// LEGACY HUB TEXT STYLES - DEPRECATED IN 3.0.0
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

// CONSOLE TEXT STYLES
interface SpanProps extends HTMLAttributes<HTMLSpanElement> {
  fontStyle: string;
}

const SpanText: React.FC<SpanProps> = ({
  children,
  className,
  fontStyle,
  ...rest
}) => {
  return (
    <span
      className={`${styles[fontStyle]} ${className ? className : ''}`}
      {...rest}
    >
      {children}
    </span>
  );
};

const BodyText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (props) => {
  return <SpanText fontStyle="bodyText" {...props} />;
};

const ErrorText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (props) => {
  return <SpanText fontStyle="errorText" {...props} />;
};

const HelpText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (props) => {
  return <SpanText fontStyle="helpText" {...props} />;
};

const FieldText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (props) => {
  return <SpanText fontStyle="fieldText" {...props} />;
};

const PlaceholderText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (
  props,
) => {
  return <SpanText fontStyle="placeholderText" {...props} />;
};

const CodeTextLarge: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <code
      className={`${styles.codeTextLarge} ${className ? className : ''}`}
      {...rest}
    >
      {children}
    </code>
  );
};

const CodeText: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <code
      className={`${styles.codeText} ${className ? className : ''}`}
      {...rest}
    >
      {children}
    </code>
  );
};

const CodeTextBlock: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <div className={styles.codeTextBlock}>
      <code
        className={`${styles.codeText} ${className ? className : ''}`}
        {...rest}
      >
        {children}
      </code>
    </div>
  );
};

const CaptionText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (
  props,
) => {
  return <SpanText fontStyle="captionText" {...props} />;
};

const CaptionTextSmall: React.FC<React.HTMLAttributes<HTMLSpanElement>> = (
  props,
) => {
  return <SpanText fontStyle="captionTextSmall" {...props} />;
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
  BodyText,
  ErrorText,
  HelpText,
  FieldText,
  PlaceholderText,
  CodeText,
  CodeTextLarge,
  CodeTextBlock,
  CaptionText,
  CaptionTextSmall,
};
