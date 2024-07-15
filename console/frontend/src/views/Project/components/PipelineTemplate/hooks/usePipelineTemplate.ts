import {useCallback, useEffect, useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';
import {useHistory} from 'react-router';

import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import useRenderTemplate from '@dash-frontend/hooks/useRenderTemplate';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';

import parseTemplateMetadata from '../lib/parseTemplateMetadata';
import {Template, TemplateArg, Templates} from '../templates';

// We want to have the arguments without default values appear first in the form.
export const sortArguments = (args: TemplateArg[]) =>
  args.sort((a, b) => {
    if (a.default === undefined && b.default !== undefined) {
      return -1;
    } else if (a.default !== undefined && b.default === undefined) {
      return 1;
    } else {
      return 0;
    }
  });

const usePipelineTemplate = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();

  const templates = useMemo(() => {
    const temp: Template[] = [];
    Templates.forEach((template) => {
      try {
        const metadata = parseTemplateMetadata(template);
        if (metadata) {
          temp.push({
            metadata,
            template,
          });
        }
      } catch (e) {
        // If we can't load stored template, gracefully handle.
        console.log(e);
      }
    });
    return temp;
  }, []);

  const goToDAG = useCallback(() => {
    browserHistory.push(lineageRoute({projectId}, false));
  }, [browserHistory, projectId]);

  const [stepTwo, setStepTwo] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<
    Template | undefined
  >(undefined);

  const formCtx = useForm({
    mode: 'onChange',
  });

  const {
    formState: {isValid},
  } = formCtx;

  const [error, setError] = useState<string>();
  const [fullError, setFullError] = useState<Error | null>();

  // Clears errors and state if we return to the first page
  useEffect(() => {
    if (!stepTwo) {
      formCtx.reset();
      setError(undefined);
      setFullError(undefined);
      setSelectedTemplate(undefined);
    }
  }, [formCtx, stepTwo]);

  const {createPipeline, loading: createLoading} = useCreatePipeline({
    onSuccess: () => {
      setError(undefined);
      setFullError(undefined);
      browserHistory.push(lineageRoute({projectId}));
    },
    onError: (error) => {
      setError('Unable to create pipeline');
      setFullError(error);
    },
  });

  const {renderTemplate, loading: templateLoading} = useRenderTemplate({
    onSuccess: (data) => {
      setError(undefined);
      setFullError(undefined);

      try {
        let jsonData = JSON.parse(data.json || '');
        // We need to add the project to the template if it is not specified
        if (!jsonData?.pipeline?.project?.name) {
          jsonData = {
            ...jsonData,
            pipeline: {...jsonData?.pipeline, project: {name: projectId}},
          };
          createPipeline({
            createPipelineRequestJson: JSON.stringify(jsonData),
          });
        }
      } catch {
        setError('Unable to read pipeline spec');
        setFullError(
          Error('Unable to parse rendered pipeline spec from pachyderm.'),
        );
      }
    },
    onError: (error) => {
      setError('Unable to parse template');
      setFullError(error);
    },
  });

  const handleSubmit = async () => {
    setError(undefined);
    setFullError(undefined);
    const values = formCtx.getValues();
    renderTemplate({
      template: selectedTemplate?.template,
      args: values,
    });
  };

  return {
    goToDAG,
    stepTwo,
    setStepTwo,
    selectedTemplate,
    setSelectedTemplate,
    formCtx,
    isValid,
    handleSubmit,
    error,
    fullError,
    loading: templateLoading || createLoading,
    templates,
  };
};

export default usePipelineTemplate;
