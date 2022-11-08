import React from 'react';

import {Story} from '@pachyderm/components';

import CreateSecondPipelineTask from './tasks/CreateMontagePipelineTask';
import MinimizeTask from './tasks/MinimizeTask';

const EasyToUseDAGs: Story = {
  name: 'Easy-to-use Directed Acyclic Graphs',
  sections: [
    {
      header: 'Add a pipeline called "montage" that uses "edges" as its input',
      taskName: 'Create montage pipeline',
      info: (
        <>
          <p>
            With Pachyderm, it&apos;s easy to create a complex DAG with
            versioning at each stage by just specifying pipelines that use
            output repos as inputs.
          </p>
          <p>
            We&apos;ll add a second pipeline, creating a 2-stage DAG. The
            montage pipeline will take the edges pipeline and the images repo as
            its inputs and stitch together the original input image and the
            edge-detected image side-by-side into a single montage.
          </p>
          <p>
            You can see how simply you can create pipeline that combines inputs
            below.
          </p>
        </>
      ),
      Task: CreateSecondPipelineTask,
    },
    {
      header: 'Pachyderm automatically processed the data',
      taskName: 'View the files in the montage repo',
      info: (
        <p>
          Once the &quot;montage&quot; pipeline was created, it was
          automatically triggered to process the data in the &quot;images&quot;
          and &quot;edges&quot; input repos, without reprocessing the data in
          the edges pipeline. You can view the files that were input and the
          processed data.
        </p>
      ),
      followUp: (
        <p>
          Adding or removing downstream pipelines takes one command in
          Pachyderm. And data existing in the output repos of upstream pipelines
          doesn&apos;t get reprocessed. Pachyderm automatically triggers
          processing of data, saving you time.
        </p>
      ),
      Task: MinimizeTask,
    },
  ],
};

export default EasyToUseDAGs;
