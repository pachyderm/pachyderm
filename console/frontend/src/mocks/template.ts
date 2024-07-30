import {rest, DefaultBodyType} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  RenderTemplateRequest,
  RenderTemplateResponse,
} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';

export const mockRenderTemplate = () =>
  rest.post<RenderTemplateRequest, Empty, RenderTemplateResponse>(
    '/api/pps_v2.API/RenderTemplate',
    (_req, res, ctx) => {
      return res(ctx.json({json: '{}'}));
    },
  );

export const mockRenderTemplateError = () =>
  rest.post<RenderTemplateRequest, Empty, RequestError>(
    '/api/pps_v2.API/RenderTemplate',
    (_req, res, ctx) => {
      return res(
        ctx.status(500),
        ctx.json({
          code: 2,
          message:
            'template err: main:69:7-12 Expected token OPERATOR but got (STRING_DOUBLE, "cmd")\n\n  "cmd":["sh"],\n\n',
          details: [],
        }),
      );
    },
  );

export const mockTextTemplate = () =>
  rest.get<DefaultBodyType, Empty, string>(
    'https://www.hosted-template-text.com',
    (_req, res, ctx) => {
      return res(ctx.body(`hello`));
    },
  );

export const mockBadFormatTemplate = () =>
  rest.get<DefaultBodyType, Empty, string>(
    'https://www.hosted-template-bad-format.com',
    (_req, res, ctx) => {
      return res(
        ctx.body(`/*
        description: "Template description."
        args:
        - description: The name of the pipeline.
          type: string
        */`),
      );
    },
  );

export const mockSimpleTemplate = () =>
  rest.get<DefaultBodyType, Empty, string>(
    'https://www.hosted-template-simple.com',
    (_req, res, ctx) => {
      return res(
        ctx.body(`/*
        title: Test Template
        description: "Template description."
        args:
        - name: name
          description: The name of the pipeline.
          default: pipeline-1
          type: string
        */`),
      );
    },
  );

export const mockSnowflakeTemplate = () =>
  rest.get<DefaultBodyType, Empty, string>(
    'https://www.hosted-template-snowflake.com',
    (_req, res, ctx) => {
      return res(
        ctx.body(`/*
        title: Snowflake Integration
        description: "Creates a cron pipeline that can execute a query against a Snowflake database and return the results in a single output file."
        args:
        - name: name
          description: The name of the pipeline.
          type: string
        - name: inputrepo
          description: The suffix to use on the cron pipeline's input repo.
          type: string
          default: tick
        - name: spec
          description: The cron spec to use. 
          type: string
          default: "@never"
        - name: query
          description: The query to execute.
          type: string
        - name: count
          description: Number input field.
          type: number
          default: 4
        */
        `),
      );
    },
  );

export const mockTemplatError = () =>
  rest.get<DefaultBodyType, Empty, string>(
    'https://www.hosted-unknown-error.com',
    (_req, res, ctx) => {
      return res(ctx.status(400));
    },
  );
