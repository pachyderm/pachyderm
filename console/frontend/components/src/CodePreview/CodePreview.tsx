import hljs from 'highlight.js/lib/core';
import 'highlight.js/styles/github.css';
import json from 'highlight.js/lib/languages/json';
import yaml from 'highlight.js/lib/languages/yaml';
import React, {useRef, useEffect} from 'react';

import styles from './CodePreview.module.css';

hljs.registerLanguage('json', json);
hljs.registerLanguage('yaml', yaml);

const CodePreview = ({
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => {
  const element = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (element.current) {
      hljs.highlightElement(element.current);
    }
  }, [element, children]);

  return (
    <div className={styles.code} ref={element} {...props}>
      {children}
    </div>
  );
};

export default CodePreview;
