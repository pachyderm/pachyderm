import {
  FileType,
  JobState,
  PipelineState,
  ProjectStatus,
  OriginKind,
  DatumState,
  State,
} from '@pachyderm/node-pachyderm';

import {OriginKind as GQLOriginKind} from '@graphqlTypes';

import {
  toGQLFileType,
  toGQLJobState,
  toGQLPipelineState,
  toGQLProjectStatus,
  toGQLCommitOrigin,
  toProtoCommitOrigin,
  toGQLDatumState,
  toGQLEnterpriseState,
} from '../gqlEnumMappers';

describe('gqlEnumMappers', () => {
  describe('toGQLPipelineState', () => {
    it('should not return an error for any proto pipeline state', () => {
      Object.values(PipelineState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLPipelineState(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLJobState', () => {
    it('should not return an error for any proto job state', () => {
      Object.values(JobState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLJobState(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLFileType', () => {
    it('should not return an error for any proto file type', () => {
      Object.values(FileType).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLFileType(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLProjectStatus', () => {
    it('should not return an error for any proto project status', () => {
      Object.values(ProjectStatus).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLProjectStatus(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLCommitOrigin', () => {
    it('should not return an error for any proto origin kind', () => {
      Object.values(OriginKind).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLCommitOrigin(val)).not.toThrowError();
      });
    });
  });

  describe('toProtoCommitOrigin', () => {
    it('should not return an error for any GQL origin', () => {
      Object.values(GQLOriginKind).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toProtoCommitOrigin(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLDatumState', () => {
    it('should not return an error for any proto datum state', () => {
      Object.values(DatumState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLDatumState(val)).not.toThrowError();
      });
    });
  });

  describe('toGQLEnterpriseState', () => {
    it('should not return an error for any enterprise state', () => {
      Object.values(State).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLEnterpriseState(val)).not.toThrowError();
      });
    });
  });
});
