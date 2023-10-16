import uniqueId from 'lodash/uniqueId';

import {FILE_UPLOAD_EXPIRATION_TIMEOUT} from './constants';

export type upload = {
  fileSets: Record<string, {id: string; deleted: boolean}>;
  path: string;
  repo: string;
  branch: string;
  description: string;
  expiration: number;
  projectId: string;
};

class FileUploads {
  private fileUploads: Record<string, upload> = {};
  private fileUploadInterval = setInterval(() => {
    for (const uploadId in this.fileUploads) {
      if (this.fileUploads[uploadId].expiration <= Date.now()) {
        const {[uploadId]: _temp, ...rest} = this.fileUploads;
        this.fileUploads = rest;
      }
    }
  }, FILE_UPLOAD_EXPIRATION_TIMEOUT);

  getUpload(uploadId: string) {
    return this.fileUploads[uploadId];
  }

  addUpload(
    path: string,
    repo: string,
    branch: string,
    description: string,
    projectId: string,
  ) {
    const id = uniqueId();

    this.fileUploads[id] = {
      fileSets: {},
      path: path,
      repo: repo,
      branch: branch,
      description: description,
      expiration: Date.now() + FILE_UPLOAD_EXPIRATION_TIMEOUT,
      projectId,
    };

    return id;
  }

  addFile(fileId: string, uploadId: string, fileSetId: string) {
    if (!this.fileUploads[uploadId].fileSets[fileId]) {
      this.fileUploads[uploadId].fileSets[fileId] = {
        id: fileSetId,
        deleted: false,
      };
    } else this.fileUploads[uploadId].fileSets[fileId].id = fileSetId;
  }

  removeUpload(uploadId: string) {
    const {[uploadId]: _temp, ...rest} = this.fileUploads;
    this.fileUploads = rest;
  }

  removeFile(fileId: string, uploadId: string) {
    if (!this.fileUploads[uploadId].fileSets[fileId]) {
      this.fileUploads[uploadId].fileSets[fileId] = {
        id: '',
        deleted: true,
      };
    } else this.fileUploads[uploadId].fileSets[fileId].deleted = true;
  }

  extendUploadExpiration(uploadId: string) {
    this.fileUploads[uploadId].expiration =
      Date.now() + FILE_UPLOAD_EXPIRATION_TIMEOUT;
  }

  clearInterval() {
    clearInterval(this.fileUploadInterval);
  }

  getFileSetIds(uploadId: string) {
    return Object.values(this.fileUploads[uploadId].fileSets).reduce<string[]>(
      (acc, fileSet) => {
        const {deleted, id} = fileSet;
        if (!deleted) acc.push(id);
        return acc;
      },
      [],
    );
  }
}

export default new FileUploads();
