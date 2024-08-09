import {EditorState, Extension} from '@codemirror/state';
import {EditorView} from '@codemirror/view';
import {githubLight} from '@uiw/codemirror-theme-github';
import classNames from 'classnames';
import {basicSetup} from 'codemirror';
import React, {useEffect, useRef} from 'react';

import {LoadingDots} from '@pachyderm/components';

import styles from './CodePreview.module.css';
import useCodePreview from './hooks/useCodePreview';
import {SupportedLanguage, getFileLanguagePlugin} from './utils/getFileDetails';

export type CodePreviewProps = {
  className?: string;
  downloadLink?: string;
  fullHeight?: boolean;
  hideGutter?: boolean;
  hideLineNumbers?: boolean;
  language?: SupportedLanguage;
  source?: string;
  sourceLoading?: boolean;
  wrapText?: boolean;
  additionalExtensions?: Extension[];
};

const CodePreview: React.FC<CodePreviewProps> = ({
  className,
  downloadLink,
  fullHeight = false,
  hideGutter = false,
  hideLineNumbers = false,
  language = 'text',
  source,
  sourceLoading,
  wrapText,
  additionalExtensions,
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
    additionalExtensions?.forEach((extension) => extensions.push(extension));

    const languagePlugin = getFileLanguagePlugin(language);

    if (languagePlugin) {
      extensions.push(languagePlugin());
    }

    if (wrapText) {
      extensions.push(EditorView.lineWrapping);
    }

    editorViewRef.current = new EditorView({
      doc: data,
      parent: codePreviewRef.current,
      extensions,
    });

    return () => {
      editorView?.destroy();
    };
  }, [data, language, wrapText, additionalExtensions]);

  if (loading || sourceLoading) {
    return (
      <div className={classNames(styles.loading, className)}>
        <LoadingDots />
      </div>
    );
  }

  return (
    <div
      className={classNames(
        styles.codePreview,
        hideGutter && styles.hideGutter,
        hideLineNumbers && styles.hideLineNumbers,
        fullHeight && styles.fullHeight,
        className,
      )}
      data-testid="CodePreview__wrapper"
    >
      <div className={styles.codePreviewContent} ref={codePreviewRef} />
    </div>
  );
};

export default CodePreview;
