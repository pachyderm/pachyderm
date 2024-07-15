import React from 'react';

import {
  ErrorText,
  HelpText,
  FieldText,
  PlaceholderText,
  CodeText,
  CodeTextLarge,
  CodeTextBlock,
  CaptionText,
  CaptionTextSmall,
} from '../Text';

import {Group} from './../../Group';
import styles from './TextStory.module.css';

export default {title: 'Text'};

type TextWrapperProps = {
  children?: React.ReactNode;
  name: string;
  family: string;
  size: string;
  lineHeight: string;
  weight: string;
  letterSpacing: string;
};

const TextWrapper: React.FC<TextWrapperProps> = ({
  name,
  family,
  size,
  lineHeight,
  weight,
  letterSpacing,
  children,
}) => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.specs}>
        <div>
          <b>Family: </b>
          {family}
        </div>
        <div>
          <b>Size: </b>
          {size}
        </div>
        <div>
          <b>Line Height: </b>
          {lineHeight}
        </div>
        <div>
          <b>Weight: </b>
          {weight}
        </div>
        <div>
          <b>Letter Spacing: </b>
          {letterSpacing}
        </div>
      </div>
      <div className={styles.content}>
        <h6 className={styles.name}>{name}</h6>
        <div className={styles.body}>{children}</div>
      </div>
    </div>
  );
};

export const Text = () => {
  return (
    <Group spacing={32} vertical>
      <TextWrapper
        name="Default Body Text"
        family="Public Sans"
        size="14px / 0.875rem"
        lineHeight="23px / 1.437rem"
        weight="300 / Light"
        letterSpacing="0.2px"
      >
        This text is applied to all elements and is the default body text within
        all pachyderm applications. It should not be modified in any way except
        for color, and the use of <b>b</b> or <strong>strong</strong> tags to{' '}
        <b>emphasize content</b>.
      </TextWrapper>
      <TextWrapper
        name="H1"
        family="Montserrat"
        size="40px / 2.5rem"
        lineHeight="50px / 3.125rem"
        weight="800 / Extra Bold"
        letterSpacing="0px"
      >
        <h1>h1 Header - Marketing copy</h1>
      </TextWrapper>
      <TextWrapper
        name="H2"
        family="Montserrat"
        size="30px / 1.875rem"
        lineHeight="40px / 2.5rem"
        weight="800 / Extra Bold"
        letterSpacing="0px"
      >
        <h2>h2 Header - Page Titles</h2>
      </TextWrapper>
      <TextWrapper
        name="H3"
        family="Montserrat"
        size="26px / 1.625rem"
        lineHeight="26px / 3.125rem"
        weight="600 / Semi Bold"
        letterSpacing="0px"
      >
        <h3>h3 Header - Emphasize Numbers</h3>
      </TextWrapper>
      <TextWrapper
        name="H4"
        family="Montserrat"
        size="18px / 1.125rem"
        lineHeight="28px / 1.75rem"
        weight="800 / Extra Bold"
        letterSpacing="0px"
      >
        <h4>h4 Header - Card / Tile Headers</h4>
      </TextWrapper>
      <TextWrapper
        name="H5"
        family="Montserrat"
        size="16px / 1rem"
        lineHeight="26px / 1.625rem"
        weight="700 / Bold"
        letterSpacing="0px"
      >
        <h5>h5 Subtitle - Card / Tile Headers</h5>
      </TextWrapper>
      <TextWrapper
        name="H6"
        family="Montserrat"
        size="14px / 0.875rem"
        lineHeight="24px / 1.5rem"
        weight="700 / Bold"
        letterSpacing="0px"
      >
        <h6>h6 Subtitle - Card / Tile Headers</h6>
      </TextWrapper>
    </Group>
  );
};

export const Special = () => {
  return (
    <Group spacing={32} vertical>
      <TextWrapper
        name="Caption Text"
        family="Montserrat"
        size="14px / 0.875rem"
        lineHeight="1.5rem"
        weight="400 / Regular"
        letterSpacing="0px"
      >
        <CaptionText>Caption Text - labels. Can be used in black.</CaptionText>
      </TextWrapper>
      <TextWrapper
        name="Caption Text Small"
        family="Montserrat"
        size="12px / 0.75rem"
        lineHeight="1.375rem"
        weight="400 / Regular"
        letterSpacing="0px"
      >
        <CaptionTextSmall>
          Caption Text Small - labels. Can be used in black.
        </CaptionTextSmall>
      </TextWrapper>
      <TextWrapper
        name="Error Text"
        family="Public Sans"
        size="14px / 0.875rem"
        lineHeight="23px 1.4375rem"
        weight="300 / Light"
        letterSpacing="0px"
      >
        <ErrorText>Error Text - this is for inline errors</ErrorText>
      </TextWrapper>
      <TextWrapper
        name="Field Text"
        family="Montserrat"
        size="12px / 0.75rem"
        lineHeight="normal"
        weight="700 / Bold"
        letterSpacing="0px"
      >
        <FieldText>
          Field Text - this is for labels above input fields
        </FieldText>
      </TextWrapper>
      <TextWrapper
        name="Help Text"
        family="Public Sans"
        size="14px / 0.875rem"
        lineHeight="23px 1.4375rem"
        weight="300 / Light"
        letterSpacing="0.2px"
      >
        <HelpText>Help Text - this is for field labels in components</HelpText>
      </TextWrapper>
      <TextWrapper
        name="Placeholder Text"
        family="Public Sans"
        size="14px / 0.875rem"
        lineHeight="23px 1.4375rem"
        weight="300 / Light"
        letterSpacing="0.2px"
      >
        <PlaceholderText>
          Placeholder Text - this is for placeholder text
        </PlaceholderText>
      </TextWrapper>
    </Group>
  );
};
export const Code = () => {
  return (
    <Group spacing={32} vertical>
      <TextWrapper
        name="Code Text Large"
        family="Menlo"
        size="14px / 0.875rem"
        lineHeight="20px 1.25rem"
        weight="400 / Regular"
        letterSpacing="0.32px"
      >
        <CodeTextLarge>
          Code text large - This is for large code snippets and larger code
          elements
        </CodeTextLarge>
      </TextWrapper>
      <TextWrapper
        name="Code Text Small"
        family="Menlo"
        size="12px / 0.75rem"
        lineHeight="16px 1rem"
        weight="400 / Regular"
        letterSpacing="0.24px"
      >
        <CodeText>
          Code Text Small - this is for inline code snippets and smaller code
          elements
        </CodeText>
      </TextWrapper>
      <TextWrapper
        name="Code Text Block"
        family="Menlo"
        size="12px / 0.75rem"
        lineHeight="16px 1rem"
        weight="400 / Regular"
        letterSpacing="0.24px"
      >
        <CodeTextBlock>
          Code Text Block - this is for inline code snippets and smaller code
          elements that include a background
        </CodeTextBlock>
      </TextWrapper>
    </Group>
  );
};
