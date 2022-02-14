import {
  FullPageModal,
  Group,
  Description,
  Form,
  Label,
  Input,
  Button,
  ButtonLink,
  Select,
  UploadSVG,
  Icon,
} from '@pachyderm/components';
import React from 'react';
import {Helmet} from 'react-helmet';

import UploadInfo from './components/UploadInfo';
import styles from './FileUpload.module.css';
import useFileUpload from './hooks/useFileUpload';

const FileUpload: React.FC = () => {
  const command =
    'pachctl put file <repo>@<branch-or-commit>[:<path/to/file>] [flags]';
  const {
    formCtx,
    isOpen,
    onClose,
    onSubmit,
    fileDrag,
    loading,
    isValid,
    fileError,
    file,
    branches,
    handleFileCancel,
    fileRegister,
    onChangeHandler,
    success,
    uploadError,
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
              <h4 className={styles.header}>Upload File</h4>
              <Description
                className={styles.terminal}
                asListItem={false}
                id="upload-file"
                copyText={command}
                title="If you have more than 1 file or a file > 1 GB, please upload via terminal:"
              >
                {command}
              </Description>
              <Group spacing={16} vertical>
                <div className={styles.prompt}>
                  If you only have 1 file &lt;1 GB:
                </div>
                <div>
                  <Label htmlFor="path" label="File Path" />
                  <Input
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
                <Label htmlFor="file" label="Attach File" />
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
                    name="file"
                    id="file"
                    disabled={loading}
                    aria-invalid={Boolean(fileError)}
                    aria-describedby="file-error"
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
                      htmlFor="file"
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
              </div>
            </Group>
            <UploadInfo
              file={file}
              loading={loading}
              uploadError={uploadError}
              fileError={fileError}
              success={success}
              handleFileCancel={handleFileCancel}
            />
          </Group>
          <div className={styles.footer}>
            <Group spacing={32} align="center">
              <span className={styles.error}>{uploadError}</span>
              <ButtonLink type="button" onClick={onClose} disabled={loading}>
                Cancel
              </ButtonLink>
              <Button disabled={!file || loading || !isValid} type="submit">
                Upload
              </Button>
            </Group>
          </div>
        </Form>
      </FullPageModal>
    </>
  );
};
export default FileUpload;
