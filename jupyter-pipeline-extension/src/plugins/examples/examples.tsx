import {JupyterFrontEnd} from '@jupyterlab/application';
import {URLExt} from '@jupyterlab/coreutils';
import {ILauncher} from '@jupyterlab/launcher';
import {ServerConnection} from '@jupyterlab/services';

import {fileIcon} from '../../utils/icons';
const EXAMPLES_CATAGORY = 'Pachyderm Examples';

export type ExampleItem = {
  title: string;
  path: string;
  command: string;
  rank: number;
};

const examples: ExampleItem[] = [
  {
    path: 'examples/Intro to Pachyderm Tutorial.ipynb',
    title: 'Intro to Pachyderm',
    command: 'jupyterlab-pachyderm:open-example-intro',
    rank: 1,
  },
  {
    path: 'examples/Mounting Data Repos in Notebooks.ipynb',
    title: 'Mounting Data Repos',
    command: 'jupyterlab-pachyderm:open-example-mount',
    rank: 2,
  },
];

const doesPathExist = async (path: string) => {
  try {
    const settings = ServerConnection.makeSettings();
    const response = await fetch(
      URLExt.join(settings.baseUrl, 'api/contents', path),
    );
    return response.status === 200;
  } catch {
    return false;
  }
};

const addExample = async (
  example: ExampleItem,
  app: JupyterFrontEnd,
  launcher: ILauncher,
) => {
  if (await doesPathExist(example.path)) {
    app.commands.addCommand(example.command, {
      icon: fileIcon,
      label: example.title,
      execute: () => {
        return app.commands.execute('docmanager:open', {
          path: example.path,
        });
      },
    });
    launcher.add({
      command: example.command,
      category: EXAMPLES_CATAGORY,
      rank: example.rank,
    });
  }
};

export const init = async (
  app: JupyterFrontEnd,
  launcher: ILauncher,
): Promise<void> => {
  await Promise.all(
    examples.map((example) => {
      return addExample(example, app, launcher);
    }),
  );
};
