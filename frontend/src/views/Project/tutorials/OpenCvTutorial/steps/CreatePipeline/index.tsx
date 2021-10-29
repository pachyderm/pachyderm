import {Step, Mark, Link} from '@pachyderm/components';
import React from 'react';

import CreatePipelineTask from './tasks/CreatePipelineTask';
import MinimizeTask from './tasks/MinimizeTask';

const CreatePipelineStep: Step = {
  label: 'Introduction',
  name: 'Creating Pipelines',
  tasks: [
    {
      name: (
        <div>
          Create a pipeline using the provided <Mark>edges.json</Mark>{' '}
          specification.
        </div>
      ),
      info: {
        name: 'Quickly define pipelines',
        text: [
          'Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo.',
        ],
      },
      Task: CreatePipelineTask,
      followUp: (
        <p>
          The pipeline spec contains a few simple sections. The pipeline section
          contains a name, which is how you will identify your pipeline. Your
          pipeline will also automatically create an output repo with the same
          name. The transform section allows you to specify the docker image you
          want to use. In this case, pachyderm/opencv is the docker image
          (defaults to DockerHub as the registry), and the entry point is{' '}
          <Mark>edges.py</Mark> The input section specifies repos visible to the
          running pipeline.
        </p>
      ),
    },
    {
      name: 'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
      Task: MinimizeTask,
      followUp:
        'The pipeline spec contains a few simple sections. The pipeline section contains a name, which is how you will identify your pipeline. Your pipeline will also automatically create an output repo with the same name. The transform section allows you to specify the docker image you want to use. In this case, pachyderm/opencv is the docker image (defaults to DockerHub as the registry), and the entry point is',
    },
  ],
  instructionsHeader: 'Creating pipelines with just a few lines of code',
  instructionsText: (
    <>
      <p>
        Pipelines are the core processing primitive in Pachyderm. Pipelines are
        defined with a simple JSON file called a pipeline specification or
        pipeline spec for short. For this example, we already created the
        pipeline spec for you. When you want to create your own pipeline specs
        later, you can refer to the full{' '}
        <Link
          externalLink
          inline
          to="https://docs.pachyderm.com/latest/reference/pipeline_spec/"
        >
          Pipeline Specification
        </Link>{' '}
        to use more advanced options. Options include building your own code
        into a container. In this tutorial, we are using a pre-built Docker
        image.
      </p>
      <p>
        For now, we are going to create a single pipeline spec that takes in
        images and does some simple edge detection.
      </p>
    </>
  ),
};

export default CreatePipelineStep;
