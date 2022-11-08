import React from 'react';

import {Story} from '@pachyderm/components';

import AddImagesTask from './tasks/AddImagesTask';
import MinimizeTask from './tasks/MinimizeTask';

const EasyToUseVersionedData: Story = {
  name: 'Easy-to-use Versioned Data',
  sections: [
    {
      header: 'Add files to the images repo you created',
      info: (
        <p>
          Now that we&apos;ve created the repo. we can add some files to it.
          Select some of the images below and click the button to add them to
          the repo.
        </p>
      ),
      taskName: 'Add files to the images repo',
      Task: AddImagesTask,
      followUp:
        "In the next step you'll view the versioned, processed data in each repo.",
    },
    {
      header: 'Pachyderm automatically processed and versioned the data.',
      info: 'Once the "edges" pipeline was created, it was automatically triggered to process the data in the "images" input repo. You can view the files that were input and the processed data.',
      taskName: 'View the files in the images and edges repos',
      Task: MinimizeTask,
      followUp: (
        <>
          <p>
            You can view these files in the images repo using the Pachyderm
            Console, as we did here, or from your computer using the pachctl
            command after connecting and authenticating to the Pachyderm
            cluster.
          </p>
          <p>
            You could also use pachctl or the Pachyderm S3 gateway to put data
            into any Pachyderm repo from any source: an S3 bucket, a url
            pointing to a website, your computer, the local filesystem of any
            computer where you can install the pachctl executable, or any url
            accessible to the Pachyderm cluster.
          </p>
          <p>
            Pachyderm can also output your pipeline&apos;s results to any
            third-party tool that uses S3 to access its data, like LabelStudio
            or Seldon, or be the data source for those tools via its S3 gateway.
          </p>
          <p>
            In the next section, we&apos;ll create a DAG by adding another
            pipeline that uses the edges output as its input.
          </p>
        </>
      ),
    },
  ],
};

export default EasyToUseVersionedData;
