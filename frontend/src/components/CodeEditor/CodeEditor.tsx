/* eslint-disable @typescript-eslint/no-explicit-any */
import {indentWithTab} from '@codemirror/commands';
import {json} from '@codemirror/lang-json';
import {EditorView, ViewUpdate, keymap} from '@codemirror/view';
import {githubLight} from '@uiw/codemirror-theme-github';
import classNames from 'classnames';
import {basicSetup} from 'codemirror';
import {jsonSchema} from 'codemirror-json-schema';
import {type JSONSchema7} from 'json-schema';
import React, {useEffect, useRef} from 'react';

import {LoadingDots} from '@pachyderm/components';

import {closeToolTipsOnType} from '../CodePreview/extensions/closeHoverTooltipsOnType';

import styles from './CodeEditor.module.css';
import {
  clusterDefaultsSchema as defaultClusterDefaultsSchema,
  createPipelineRequestSchema as defaultCreatePipelineRequestSchema,
} from './schemas';
import {useJsonSchema} from './useJsonSchema';

const schemaMapper: Record<
  'createPipelineRequest' | 'clusterDefaults',
  {url: string; fallback: JSONSchema7}
> = {
  clusterDefaults: {
    url: 'pps_v2/ClusterDefaults.schema.json',
    fallback: defaultClusterDefaultsSchema,
  },
  createPipelineRequest: {
    url: 'pps_v2/CreatePipelineRequest.schema.json',
    fallback: defaultCreatePipelineRequestSchema,
  },
};

type CodePreviewProps = {
  className?: string;
  downloadLink?: string;
  fullHeight?: boolean;
  hideGutter?: boolean;
  hideLineNumbers?: boolean;
  language?: 'json' | 'markdown' | 'yaml' | 'text';
  source?: string;
  initialDoc?: string;
  loading: boolean;
  onChange?: (val: string) => void;
  schema?: 'createPipelineRequest' | 'clusterDefaults';
  dataTestid?: string;
};
const CodeEditor: React.FC<CodePreviewProps> = ({
  className,
  fullHeight = false,
  hideGutter = false,
  hideLineNumbers = false,
  initialDoc,
  source,
  loading,
  onChange,
  schema,
  dataTestid,
}) => {
  const codeEditorRef = useRef<HTMLDivElement | null>(null);
  const editorViewRef = useRef<EditorView | null>(null);

  const {schema: jsonSchemaIn, loading: schemaLoading} = useJsonSchema(
    schemaMapper[schema || 'clusterDefaults'].url,
    schemaMapper[schema || 'clusterDefaults'].fallback,
    !schema,
  );

  useEffect(() => {
    if (!codeEditorRef.current) return;

    const editorView = editorViewRef.current;

    if (editorView) {
      editorView.destroy();
    }

    const extensions = [
      basicSetup,
      githubLight,
      json(),
      keymap.of([indentWithTab]),
      EditorView.updateListener.of((view: ViewUpdate) => {
        if (view.docChanged) {
          onChange && onChange(view.state.doc.toString());
        }
      }),
      closeToolTipsOnType(),
    ];

    if (schema && jsonSchemaIn) extensions.push(jsonSchema(jsonSchemaIn));

    editorViewRef.current = new EditorView({
      doc: initialDoc,
      parent: codeEditorRef.current,
      extensions,
    });

    return () => {
      editorView?.destroy();
    };
  }, [initialDoc, onChange, jsonSchemaIn, schema]);

  // update editor on source changes
  useEffect(() => {
    if (editorViewRef.current?.state.doc.toString() !== source) {
      editorViewRef.current?.dispatch({
        changes: {
          from: 0,
          to: editorViewRef.current?.state.doc.toString().length,
          insert: source,
        },
      });
    }
  }, [source]);

  if (loading || schemaLoading) {
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
      data-testid={dataTestid}
    >
      <div
        className={styles.codePreviewContent}
        ref={codeEditorRef}
        aria-label="code editor"
      />
    </div>
  );
};

export default CodeEditor;
