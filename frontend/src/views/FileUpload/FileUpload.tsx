import {
  FullPageModal,
  Group,
  Form,
  Label,
  Input,
  Button,
  ButtonLink,
  Select,
  UploadSVG,
  Icon,
  CopySVG,
} from '@pachyderm/components';
import React from 'react';
import {Helmet} from 'react-helmet';

import EmptyState from '../../components/EmptyState';

import FileCard from './components/FileCard';
import styles from './FileUpload.module.css';
import useFileUpload from './hooks/useFileUpload';
import {GLOB_CHARACTERS} from './lib/constants';

const FileUpload: React.FC = () => {
  const {
    formCtx,
    isOpen,
    onClose,
    onSubmit,
    fileDrag,
    loading,
    isValid,
    files,
    branches,
    handleFileCancel,
    fileRegister,
    onChangeHandler,
    handleCopy,
    copySupported,
    error,
    uploadId,
    maxStreamIndex,
    setMaxStreamIndex,
    onError,
    fileNameError,
    uploadsFinished,
    handleDoneClick,
    finishLoading,
  } = useFileUpload();

  return (
    <>
      <Helmet>
        <title>Upload Files - Pachyderm Console</title>
      </Helmet>
      <FullPageModal hideType="exit" show={isOpen} onHide={onClose}>
        <Form formContext={formCtx} className={styles.form} onSubmit={onSubmit}>
          <Group className={styles.base}>
            <Group spacing={32} vertical className={styles.fileForm}>
              <Group vertical spacing={8} align="start">
                <ButtonLink
                  className={styles.terminal}
                  onClick={handleCopy}
                  disabled={!copySupported}
                >
                  Upload via terminal
                  <Icon className={styles.copySVG} color="plum" small>
                    <CopySVG />
                  </Icon>
                </ButtonLink>
                <h4 className={styles.header}>Upload Files</h4>
              </Group>
              <Group spacing={16} vertical>
                <div>
                  <Label htmlFor="path" label="File Path" />
                  <Input
                    disabled={loading}
                    type="text"
                    id="path"
                    name="path"
                    defaultValue="/"
                    validationOptions={{
                      pattern: {
                        value: /^\/[\w-/]*$/g,
                        message:
                          'Paths can only contain alphanumeric characters and must start with a forward slash',
                      },
                      required: {
                        value: true,
                        message: 'A path is required',
                      },
                    }}
                  />
                </div>
              </Group>
              <div>
                <Label htmlFor="branch" label="Branch" />
                <Select id="branch" initialValue="master">
                  {branches.map((branch) => (
                    <Select.Option value={branch} key={branch}>
                      {branch}
                    </Select.Option>
                  ))}
                </Select>
              </div>
              <div>
                <Label htmlFor="files" label="Attach Files" />
                <div
                  className={styles.fileUpload}
                  onDrop={fileDrag}
                  onDragOver={(e) => e.preventDefault()}
                >
                  <input
                    onChange={onChangeHandler}
                    {...fileRegister}
                    className={styles.fileUploadInput}
                    type="file"
                    name="files"
                    id="files"
                    disabled={loading}
                    aria-invalid={Boolean(fileNameError)}
                    aria-describedby="file-error"
                    multiple
                  />
                  <Group vertical spacing={16}>
                    <Group spacing={8} justify="center">
                      <Icon disabled={loading}>
                        <UploadSVG />
                      </Icon>
                      <span
                        className={`${styles.fileUploadText}
                        ${loading ? styles.fileUploadTextDisabled : ''}`}
                      >
                        Drag and drop file here
                      </span>
                    </Group>
                    <span
                      className={`${styles.fileUploadText}
                        ${loading ? styles.fileUploadTextDisabled : ''}`}
                    >
                      Or
                    </span>
                    <label
                      htmlFor="files"
                      onClick={(e) => {
                        if (e.target !== e.currentTarget)
                          e.currentTarget.click();
                      }}
                    >
                      <Button
                        buttonType="secondary"
                        autoSize
                        disabled={loading}
                        className={styles.fileButton}
                        type="button"
                      >
                        Browse Files
                      </Button>
                    </label>
                  </Group>
                </div>
                <div className={styles.uploadHelp}>
                  <em>We cannot accept {GLOB_CHARACTERS} in file names.</em>
                </div>
              </div>
            </Group>
            <Group spacing={8} vertical className={styles.uploadInfo}>
              <div className={styles.uploadInfoText}>
                {(fileNameError || error) && (
                  <span className={styles.error} id="file-error">
                    {fileNameError?.message || error}
                  </span>
                )}
                {uploadsFinished && (
                  <em className={styles.uploadSucessText}>
                    All files have been successfully uploaded, click Done to
                    commit the files.
                  </em>
                )}
              </div>
              <div>
                {files.length > 0 ? (
                  <Group vertical spacing={8} className={styles.fileCards}>
                    {files.map((file, i) => {
                      return (
                        <FileCard
                          uploadId={uploadId}
                          file={file}
                          handleFileCancel={handleFileCancel}
                          key={file.name}
                          index={i}
                          maxStreamIndex={maxStreamIndex}
                          onComplete={setMaxStreamIndex}
                          onError={onError}
                          uploadError={Boolean(error)}
                        />
                      );
                    })}
                  </Group>
                ) : (
                  <Group align="start">
                    <EmptyState
                      title="Let's Start"
                      message="Upload your file on the left!"
                      className={styles.emptyState}
                    />
                  </Group>
                )}
              </div>
            </Group>
          </Group>
          <div className={styles.footer}>
            <Group spacing={32} align="center">
              <ButtonLink type="button" onClick={onClose}>
                Cancel
              </ButtonLink>
              {uploadsFinished ? (
                <Button
                  type="button"
                  onClick={handleDoneClick}
                  disabled={finishLoading}
                >
                  Done
                </Button>
              ) : (
                <Button
                  disabled={files.length === 0 || loading || !isValid}
                  type="submit"
                >
                  Upload
                </Button>
              )}
            </Group>
          </div>
        </Form>
      </FullPageModal>
    </>
  );
};
export default FileUpload;
