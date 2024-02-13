import classnames from 'classnames';
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

const IdText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <span
      className={`${styles.idText} ${className ? className : ''}`}
      {...rest}
    >
      {children}
    </span>
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

const CaptionColors = {
  white: 'white',
  black: 'black',
  grey: 'captionGrey',
};

const defaultCaptionColor = CaptionColors.grey;

type CaptionColor = keyof typeof CaptionColors;

export type CaptionProps = React.HTMLAttributes<HTMLSpanElement> & {
  color?: CaptionColor;
};

const CaptionText: React.FC<CaptionProps> = ({color, className, ...rest}) => {
  const actualColor = (color && CaptionColors[color]) || defaultCaptionColor;
  return (
    <SpanText
      fontStyle="captionText"
      className={classnames(className, styles[actualColor])}
      {...rest}
    />
  );
};

const CaptionTextSmall: React.FC<CaptionProps> = ({
  color,
  className,
  ...rest
}) => {
  const actualColor = (color && CaptionColors[color]) || defaultCaptionColor;
  return (
    <SpanText
      fontStyle="captionTextSmall"
      className={classnames(className, styles[actualColor])}
      {...rest}
    />
  );
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
  IdText,
  CaptionText,
  CaptionTextSmall,
};
