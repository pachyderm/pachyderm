import {json} from '@codemirror/lang-json';
import {markdown} from '@codemirror/lang-markdown';
import {EditorState} from '@codemirror/state';
import {githubLight} from '@uiw/codemirror-theme-github';
import classNames from 'classnames';
import {EditorView, basicSetup} from 'codemirror';
import React, {useEffect, useRef} from 'react';

import {SupportedLanguage} from '@dash-frontend/lib/getFileDetails';
import {LoadingDots} from '@pachyderm/components';

import styles from './CodePreview.module.css';
import yaml from './extensions/yaml';
import useCodePreview from './hooks/useCodePreview';

type CodePreviewProps = {
  className?: string;
  downloadLink?: string;
  fullHeight?: boolean;
  hideGutter?: boolean;
  hideLineNumbers?: boolean;
  language?: SupportedLanguage;
  source?: string;
};

const CodePreview: React.FC<CodePreviewProps> = ({
  className,
  downloadLink,
  fullHeight = false,
  hideGutter = false,
  hideLineNumbers = false,
  language = 'text',
  source,
}) => {
  const codePreviewRef = useRef<HTMLDivElement | null>(null);
  const editorViewRef = useRef<EditorView | null>(null);
  const {data, loading} = useCodePreview(downloadLink, source);

  useEffect(() => {
    if (!codePreviewRef.current) return;

    const editorView = editorViewRef.current;

    if (editorView) {
      editorView.destroy();
    }

    const extensions = [basicSetup, githubLight, EditorState.readOnly.of(true)];

    if (language === 'json') {
      extensions.push(json());
    } else if (language === 'markdown') {
      extensions.push(markdown());
    } else if (language === 'yaml') {
      extensions.push(yaml());
    }

    editorViewRef.current = new EditorView({
      doc: data,
      parent: codePreviewRef.current,
      extensions,
    });

    return () => {
      editorView?.destroy();
    };
  }, [data, language]);

  if (loading) return <LoadingDots />;

  return (
    <div
      className={classNames(
        className,
        styles.codePreview,
        hideGutter && styles.hideGutter,
        hideLineNumbers && styles.hideLineNumbers,
        fullHeight && styles.fullHeight,
      )}
      data-testid="CodePreview__wrapper"
    >
      <div className={styles.codePreviewContent} ref={codePreviewRef} />
    </div>
  );
};

export default CodePreview;
