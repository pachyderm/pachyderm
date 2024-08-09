import {useMutation} from '@tanstack/react-query';

import {
  RenderTemplateRequest,
  RenderTemplateResponse,
  renderTemplate,
} from '@dash-frontend/api/pps';

type useRenderTemplateArgs = {
  onSuccess?: (data: RenderTemplateResponse) => void;
  onError?: (error: Error) => void;
};

const useRenderTemplate = ({onSuccess, onError}: useRenderTemplateArgs) => {
  const {
    mutate: renderTemplateMutation,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['renderTemplate'],
    mutationFn: (req: RenderTemplateRequest) => renderTemplate(req),
    onSuccess: (data: RenderTemplateResponse) => {
      onSuccess && onSuccess(data);
    },
    onError: (error) => onError && onError(error),
  });

  return {
    renderTemplateResponse: data,
    renderTemplate: renderTemplateMutation,
    error,
    loading,
  };
};

export default useRenderTemplate;
