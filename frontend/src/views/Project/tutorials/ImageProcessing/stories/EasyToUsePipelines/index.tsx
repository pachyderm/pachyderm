import React from 'react';

import {Story} from '@pachyderm/components';

import CreatePipelineTask from './tasks/CreatePipelineTask';
import CreateRepoTask from './tasks/CreateRepoTask';

const EasyToUsePipelines: Story = {
  name: 'Easy-to-use Pipelines',
  sections: [
    {
      header: 'Video and image processing at scale with Pachyderm',
      info: (
        <>
          <p>
            This tutorial walks you through the deployment of a two-pipeline DAG
            that processes images, performing edge detection and then creating a
            single montage of the images. We&apos;ll take you through creating
            input data repositories for storing versioned data, and two
            data-driven pipelines that automatically process data as it&apos;s
            received.
          </p>
          <p>
            Pachyderm&apos;s data-driven pipelines can be defined in just a few
            lines of JSON or YAML.You just need to define the inputs, the docker
            image, the command to be executed, and give it a name.
          </p>
          <p>
            In the next few steps, you&apos;ll create an input repository called
            <q>images</q>, define a pipeline called <q>edges</q> that takes
            input from images, and after that you will process some images with
            the <q>edges</q> pipeline.
          </p>
        </>
      ),
    },
    {
      header: 'Create the images repository',
      taskName: 'Create images repo',
      info: (
        <p>
          Before you create a Pachyderm data-driven pipeline, you need to create
          at least one versioned data repository from which it will get its
          data. In this case, the repo will be named <q>images</q>, because
          it&apos;s where we&apos;ll put our images to be processed.
        </p>
      ),
      Task: CreateRepoTask,
    },
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
            creating a data-driven pipeline using 9 lines of easily-understood
            JSON.
          </p>
          <ul>
            <li>2 lines identify the pipeline</li>
            <li>4 lines define the input</li>
            <li>3 lines specify the command</li>
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

export default EasyToUsePipelines;
