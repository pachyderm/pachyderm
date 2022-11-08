import classnames from 'classnames';
import React from 'react';

import {CodePreview} from '@pachyderm/components';

import stringify, {Format} from '../../utils/stringifyToFormat';
import MinimizableSection from '../MinimizableSection';

import styles from './CodeElement.module.css';
type Element = Record<string, unknown> | [];

interface CodeElementProps {
  element: Element;
  format: Format;
  depth?: number;
}

const CodeFragment: React.FC<{format: Format; depth: number}> = ({
  children,
  format,
  depth,
}) => {
  return (
    <CodePreview
      className={classnames(styles.code, `language-${format.toLowerCase()}`, {
        [styles[`depth${depth === 0 ? '0' : depth % 6}`]]: true,
      })}
    >
      {children}
    </CodePreview>
  );
};

const CodeElement: React.FC<CodeElementProps> = ({
  element,
  format,
  depth = 0,
}) => {
  const expandableSections: Element[] = [];
  const placeholder = {
    __placeholder: true,
  };
  let elementString = stringify(
    Object.keys(element).reduce(
      (memo: Record<string, unknown>, key) => {
        const child = Array.isArray(element)
          ? element[parseInt(key)]
          : element[key];
        if (child) {
          const expandable =
            !Array.isArray(element) &&
            typeof child === 'object' &&
            Object.keys(child).length;
          if (expandable) {
            expandableSections.push(child as Element);
            memo[key] = Array.isArray(child)
              ? [{...placeholder}]
              : {...placeholder};
          } else {
            memo[key] = child;
          }
        }
        return memo;
      },
      Array.isArray(element) ? [] : {},
    ),
    format,
  );

  if (depth && format === Format.JSON) {
    // remove open and close brackets
    elementString = elementString.slice(2, elementString.length - 2);
  }

  const placeholderString = stringify(placeholder, format).replace(
    /[{}\n+]/g,
    '',
  );
  const fragments = elementString
    .split(placeholderString + '\n')
    .map((fragment) =>
      fragment.replace('[\n    {\n  ', '[\n').replace('    }\n  ]', '  ]'),
    );

  return (
    <div className={styles.codeSection}>
      {fragments.map((fragment, index) => {
        if (expandableSections[index]) {
          const lines = fragment.split('\n');
          const headerFragment = lines[lines.length - 2];
          fragment = fragment.slice(0, fragment.indexOf(headerFragment));
          return (
            <div key={`${depth}:${index}`}>
              <CodeFragment format={format} depth={depth}>
                {fragment}
              </CodeFragment>
              <MinimizableSection
                header={
                  <CodeFragment format={format} depth={depth}>
                    {headerFragment}
                  </CodeFragment>
                }
              >
                <CodeElement
                  depth={depth + 1}
                  element={expandableSections[index]}
                  format={format}
                />
              </MinimizableSection>
            </div>
          );
        }
        return (
          <CodeFragment key={`${depth}:${index}`} format={format} depth={depth}>
            {fragment}
          </CodeFragment>
        );
      })}
    </div>
  );
};

export default CodeElement;
