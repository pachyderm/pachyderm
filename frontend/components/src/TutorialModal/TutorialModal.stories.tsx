import React, {useCallback} from 'react';

import {useMinimizeTask} from '@pachyderm/components';

import {Link} from '../Link';

import Mark from './components/Mark';
import {
  MultiSelectModule,
  useMultiSelectModule,
  ConfigurationUploadModule,
} from './components/Modules/';
import TaskCard from './components/TaskCard';
import TutorialModalBodyProvider from './components/TutorialModalBody/TutorialModalBodyProvider';
import {Story, TaskComponentProps} from './lib/types';
import TutorialModal from './TutorialModal';

export default {title: 'TutorialModal'};

const Task1Component: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const files = {
    '/birthday-cake.jpg': {
      name: 'cake.png',
      path: '/birthday-cake.jpg',
    },
    '/sutro-tower.jpg': {
      name: 'sutrotower.jpg',
      path: '/sutro-tower.jpg',
    },
    '/wine.jpg': {
      name: 'wine.jpg',
      path: '/wine.jpg',
    },
    '/puppy.png': {
      name: 'puppy.png',
      path: '/puppy.png',
    },
  };

  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});

  const action = useCallback(() => {
    setDisabled(true);
    setUploaded();
    onCompleted();
  }, [onCompleted, setDisabled, setUploaded]);

  return (
    <TaskCard
      index={index}
      task={name}
      action={action}
      currentTask={currentTask}
      actionText="Create pipeline spec"
      taskInfoTitle="Quickly define pipelines"
      taskInfo="Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo."
    >
      <MultiSelectModule repo="images" type="image" {...register} />
    </TaskCard>
  );
};

const Task2Component: React.FC<TaskComponentProps> = ({
  currentTask,
  onCompleted,
  minimized,
  index,
}) => {
  useMinimizeTask({currentTask, onCompleted, minimized, index});

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

const Task3Component: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const files = {
    '/File1.csv': {
      name: 'File1.csv',
      path: '/Vile1.csv',
    },
    '/File2.csv': {
      name: 'File2.csv',
      path: '/File2.csv',
    },
    '/File3.csv': {
      name: 'File3.csv',
      path: '/File2.csv',
    },
    '/File4.csv': {
      name: 'File4.csv',
      path: '/File4.csv',
    },
  };

  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});

  const action = useCallback(() => {
    setDisabled(true);
    setUploaded();
    onCompleted();
  }, [onCompleted, setDisabled, setUploaded]);

  return (
    <TaskCard
      index={index}
      task={name}
      action={action}
      currentTask={currentTask}
      actionText="Create pipeline spec"
      taskInfoTitle="Quickly define pipelines"
      taskInfo="Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo."
    >
      <MultiSelectModule repo="images" type="file" {...register} />
    </TaskCard>
  );
};

const stories: Story[] = [
  {
    name: 'Easy-to-use Pipelines',
    sections: [
      {
        header: 'Section without Task',
        info: 'This is a section without a task.',
      },
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
        header: 'Section without Task',
        info: 'This is a section without a task.',
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
        Task: Task3Component,
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
        header: 'Section without Task',
        info: 'This is a section without a task.',
      },
    ],
  },
  {
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
        Task: Task3Component,
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
    ],
  },
];

export const Default = () => {
  return (
    <TutorialModalBodyProvider>
      <TutorialModal stories={stories} tutorialName="image-processing" />
    </TutorialModalBodyProvider>
  );
};

const ConfigurationTaskComponent: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const action = useCallback(() => {
    onCompleted();
  }, [onCompleted]);

  const fileMeta = {
    name: 'edges.json',
    path: 'https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.pipeline.json',
  };

  const fileContent = `{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "input": {
    "pfs": {
      "repo": "images",
      "glob": "/*"
    }
  }
}`;

  return (
    <TaskCard
      index={index}
      task={name}
      action={action}
      currentTask={currentTask}
      actionText="Create pipeline spec"
      taskInfoTitle="Quickly define pipelines"
      taskInfo="Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo."
    >
      <ConfigurationUploadModule
        fileMeta={fileMeta}
        fileContents={fileContent}
      />
    </TaskCard>
  );
};

const configurationStories: Story[] = [
  {
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
        Task: ConfigurationTaskComponent,
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
    ],
  },
];

export const ConfigurationUpload = () => {
  return (
    <TutorialModalBodyProvider>
      <TutorialModal
        stories={configurationStories}
        tutorialName="configuration-upload"
        onClose={() => null}
        onSkip={() => null}
      />
    </TutorialModalBodyProvider>
  );
};
