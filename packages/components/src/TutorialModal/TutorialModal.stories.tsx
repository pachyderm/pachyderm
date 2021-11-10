import React, {useEffect} from 'react';

import {Link} from '../Link';

import Mark from './components/Mark';
import TaskCard from './components/TaskCard';
import {Step, TaskComponentProps} from './lib/types';
import TutorialModal from './TutorialModal';

export default {title: 'TutorialModal'};

const Task1Component: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  return (
    <TaskCard
      index={index}
      task={name}
      action={onCompleted}
      currentTask={currentTask}
      actionText="Create pipeline spec"
      taskInfoTitle="Quickly define pipelines"
      taskInfo="Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo."
    />
  );
};

const Task2Component: React.FC<TaskComponentProps> = ({
  currentTask,
  onCompleted,
  minimized,
  index,
}) => {
  useEffect(() => {
    if (currentTask === index && minimized) {
      onCompleted();
    }
  }, [currentTask, minimized, onCompleted, index]);

  return (
    <TaskCard
      index={index}
      currentTask={currentTask}
      task="Minimize the overlay and inspect the pipeline and resulting output repo in the DAG"
      taskInfoTitle="Automatically track data lineage"
      taskInfo="Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book."
    />
  );
};

const steps: Step[] = [
  {
    label: 'Introduction',
    name: 'Creating Pipelines',
    sections: [
      {
        header: 'Creating pipelines with just a few lines of code',
        info: (
          <>
            <p>
              Pipelines are the core processing primitive in Pachyderm.
              Pipelines are defined with a simple JSON file called a pipeline
              specification or pipeline spec for short. For this example, we
              already created the pipeline spec for you. When you want to create
              your own pipeline specs later, you can refer to the full{' '}
              <Link externalLink inline to="foo">
                Pipeline Specification
              </Link>{' '}
              to use more advanced options. Options include building your own
              code into a container. In this tutorial, we are using a pre-built
              Docker image.
            </p>
            <span>
              For now, we are going to create a single pipeline spec that takes
              in images and does some simple edge detection.
            </span>
          </>
        ),
        taskName: (
          <span>
            Create a pipeline using the provided <Mark>edges.json</Mark>{' '}
            specification.
          </span>
        ),
        Task: Task1Component,
        followUp: (
          <p>
            The pipeline spec contains a few simple sections. The pipeline
            section contains a name, which is how you will identify your
            pipeline. Your pipeline will also automatically create an output
            repo with the same name. The transform section allows you to specify
            the docker image you want to use. In this case, pachyderm/opencv is
            the docker image (defaults to DockerHub as the registry), and the
            entry point is <Mark>edges.py</Mark> The input section specifies
            repos visible to the running pipeline.
          </p>
        ),
      },
      {
        isSubHeader: true,
        header: 'Viewing commits in Console',
        info: (
          <p>
            Console provides // Lorem Ipsum is simply dummy text of the printing
            and typesetting industry. Lorem Ipsum has been the industry&apos;s
            standard dummy text ever since the 1500s, when an unknown printer
            took a galley of type and scrambled it to make a type specimen book.
            It has survived not only five centuries, but also the leap into
            electronic typesetting, remaining essentially unchanged. It was
            popularised in the 1960s with the release of Letraset sheets
            containing Lorem Ipsum passages, and more recently with desktop
            publishing software like Aldus PageMaker including versions of Lorem
            Ipsum.
          </p>
        ),
        taskName:
          'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
        Task: Task2Component,
        followUp:
          'The pipeline spec contains a few simple sections. The pipeline section contains a name, which is how you will identify your pipeline. Your pipeline will also automatically create an output repo with the same name. The transform section allows you to specify the docker image you want to use. In this case, pachyderm/opencv is the docker image (defaults to DockerHub as the registry), and the entry point is',
      },
    ],
  },
];

export const Default = () => {
  return <TutorialModal steps={steps} />;
};
