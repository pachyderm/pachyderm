import React, {HTMLAttributes} from 'react';

import styles from './Text.module.css';
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
