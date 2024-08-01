import rawClusterDefaultsSchema from '@pachyderm/src/internal/jsonschema/pps_v2/ClusterDefaults.schema.json';
import rawCreatePipelineRequestSchema from '@pachyderm/src/internal/jsonschema/pps_v2/CreatePipelineRequest.schema.json';
import rawProjectDefaultsSchema from '@pachyderm/src/internal/jsonschema/pps_v2/ProjectDefaults.schema.json';
import {rest} from 'msw';

export const mockClusterDefaultsSchema = rest.get(
  '/jsonschema/pps_v2/ClusterDefaults.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json(rawClusterDefaultsSchema));
  },
);

export const mockClusterDefaultsSchema404 = rest.get(
  '/jsonschema/pps_v2/ClusterDefaults.schema.json',
  (_, res, ctx) => res(ctx.status(404)),
);

export const mockClusterDefaultsSchemaBadData = rest.get(
  '/jsonschema/pps_v2/ClusterDefaults.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json({foo: 'bar'}));
  },
);

export const mockCreatePipelineRequestSchema = rest.get(
  '/jsonschema/pps_v2/CreatePipelineRequest.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json(rawCreatePipelineRequestSchema));
  },
);

export const mockCreatePipelineRequestSchema404 = rest.get(
  '/jsonschema/pps_v2/CreatePipelineRequest.schema.json',
  (_, res, ctx) => res(ctx.status(404)),
);

export const mockCreatePipelineRequestSchemaBadData = rest.get(
  '/jsonschema/pps_v2/CreatePipelineRequest.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json({foo: 'bar'}));
  },
);

export const mockProjectDefaultsSchema = rest.get(
  '/jsonschema/pps_v2/ProjectDefaults.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json(rawProjectDefaultsSchema));
  },
);

export const mockProjectDefaultsSchema404 = rest.get(
  '/jsonschema/pps_v2/ProjectDefaults.schema.json',
  (_, res, ctx) => res(ctx.status(404)),
);

export const mockProjectDefaultsSchemaBadData = rest.get(
  '/jsonschema/pps_v2/ProjectDefaults.schema.json',
  (_, res, ctx) => {
    return res(ctx.status(200), ctx.json({foo: 'bar'}));
  },
);
