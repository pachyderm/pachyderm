import {Story} from '@pachyderm/components';
import React from 'react';

import CreatePipelineTask from './tasks/CreatePipelineTask';

const CreatePipelineStory: Story = {
  name: 'Creating Pipelines',
  sections: [
    {
      header: 'Create a pipeline',
      taskName: <div>Create the edges pipeline</div>,
      info: (
        <>
          <p>
            Pachyderm makes creating data-driven pipelines easy and simple. You
            only need to know four things: the name you want to give the
            pipeline, the tag for the docker container you want to use, the
            command you want to execute, and the input repo or repos.
          </p>
          <p>This pipeline takes an image and performs edge detection on it.</p>
        </>
      ),
      Task: CreatePipelineTask,
      followUp: (
        <>
          <p>
            Note that there was no need to create the input queues, output
            queues, define exception handling, or even create the data
            repository if it already exists. There is just one command for
            creating a data-driven pipeline using 11 lines of easily-understood
            YAML.
          </p>
          <ul>
            <li>2 lines identify the pipeline</li>
            <li>4 lines define the input</li>
            <li>1 line specifies the input</li>
            <li>4 lines specifies the command</li>
          </ul>
          <p>
            Each pipeline automatically creates a versioned data repository to
            store its processed results. As we will see in a future step, that
            repository can, in turn, be used as an input to other pipelines,
            creating complex DAGs that provide complete versioning and
            reproducibility at each stage.
          </p>
          <p>
            Pachyderm&apos;s data-driven pipelines are easy to create either
            manually or programmatically.
          </p>
        </>
      ),
    },
  ],
};

export default CreatePipelineStory;
