import {
  FileType,
  JobState,
  PipelineState,
  OriginKind,
  DatumState,
  State,
  Permission,
} from '@dash-backend/proto';
import {
  OriginKind as GQLOriginKind,
  ResourceType as GQLResourceType,
  Permission as GQLPermission,
} from '@graphqlTypes';

import {
  toGQLFileType,
  toGQLJobState,
  toGQLPipelineState,
  toGQLCommitOrigin,
  toProtoCommitOrigin,
  toGQLDatumState,
  toGQLEnterpriseState,
  toProtoResourceType,
  toProtoPermissionType,
  toGQLPermissionType,
} from '../gqlEnumMappers';

describe('gqlEnumMappers', () => {
  describe('toGQLPipelineState', () => {
    it('should not return an error for any proto pipeline state', () => {
      Object.values(PipelineState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLPipelineState(val)).not.toThrow();
      });
    });
  });

  describe('toGQLJobState', () => {
    it('should not return an error for any proto job state', () => {
      Object.values(JobState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLJobState(val)).not.toThrow();
      });
    });
  });

  describe('toGQLFileType', () => {
    it('should not return an error for any proto file type', () => {
      Object.values(FileType).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLFileType(val)).not.toThrow();
      });
    });
  });

  describe('toGQLCommitOrigin', () => {
    it('should not return an error for any proto origin kind', () => {
      Object.values(OriginKind).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLCommitOrigin(val)).not.toThrow();
      });
    });
  });

  describe('toProtoCommitOrigin', () => {
    it('should not return an error for any GQL origin', () => {
      Object.values(GQLOriginKind).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toProtoCommitOrigin(val)).not.toThrow();
      });
    });
  });

  describe('toGQLDatumState', () => {
    it('should not return an error for any proto datum state', () => {
      Object.values(DatumState).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLDatumState(val)).not.toThrow();
      });
    });
  });

  describe('toGQLEnterpriseState', () => {
    it('should not return an error for any enterprise state', () => {
      Object.values(State).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLEnterpriseState(val)).not.toThrow();
      });
    });
  });

  describe('toProtoResourceType', () => {
    it('should not return an error for any resource type', () => {
      Object.values(GQLResourceType).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toProtoResourceType(val)).not.toThrow();
      });
    });
  });

  describe('toProtoPermissionType', () => {
    it('should not return an error for any resource type', () => {
      Object.values(GQLPermission).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toProtoPermissionType(val)).not.toThrow();
      });
    });
  });

  describe('toGQLPermissionType', () => {
    it('should not return an error for any resource type', () => {
      Object.values(Permission).forEach((val) => {
        if (typeof val === 'string') return;
        expect(() => toGQLPermissionType(val)).not.toThrow();
      });
    });
  });
});
