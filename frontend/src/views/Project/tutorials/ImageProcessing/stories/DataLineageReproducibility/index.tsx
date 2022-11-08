import React from 'react';

import {Story} from '@pachyderm/components';

import AddFilesTask from './tasks/AddFilesTask/AddFilesTask';
import BasicConceptsTask from './tasks/BasicConceptsTask';
import CheckMontageVersionTask from './tasks/CheckMontageVersionTask';
import CheckNewImagesTask from './tasks/CheckNewImagesTask';
import MoveBranchTask from './tasks/MoveBranchTask';

const DataLineageReproducibility: Story = {
  name: 'Data lineage makes reproducibility easy',
  sections: [
    {
      header: 'Basic reproducibility concepts',
      info: (
        <>
          <p>
            Pachyderm saves the results of all processing in immutable commits,
            including intermediate results.
          </p>
          <p>
            Those immutable commits contain files with deduplicated data that
            were either uploaded by the user or are the output of user code.
          </p>
          <p>
            The provenance chain starts with input commits using global
            identifiers.
          </p>
          <p>
            You can find any result, anywhere in your DAG, using that global
            identifier.
          </p>
          <p>You can reproduce any result using that global identifier, too.</p>
          <p>
            Pipelines, by default, process the head commits in the master branch
            of their input repos.
          </p>
          <p>
            Repointing the head commit in the master branch to a new global
            identifier is one way you can reproduce results through a complex
            DAG.
          </p>
          <p>
            In this section, we&apos;ll observe the current state of the montage
            file, add some files to the images repo, see how the montage
            changes, and then reproduce the initial state of the montage file.
          </p>
        </>
      ),
      taskName: 'Look at the current montage file',
      Task: BasicConceptsTask,
    },
    {
      header: 'Add files to the images repo',
      info: (
        <>
          <p>
            We&apos;ll add more files to the images repo so that the montage
            output changes.
          </p>
        </>
      ),
      taskName: 'Add files to the images repo',
      Task: AddFilesTask,
    },
    {
      header: 'Confrim that the montage has new images added',
      info: (
        <>
          <p>
            Once the job triggered by adding new images is done, you should see
            the the head commit in the montage repo&apos;s master branch updated
            with new data.
          </p>
        </>
      ),
      taskName: 'Look at the updated montage file',
      Task: CheckNewImagesTask,
    },
    {
      header: "Reproduce results through Pachyderm's immutable commits",
      info: (
        <>
          <p>
            In this section, you&apos;ll move the master branch in the images
            repo to point at the very first commit you made in this tutorial.
            You&apos;ll see that the results in the montage repo are quickly and
            easily reproduced.
          </p>
        </>
      ),
      taskName: "Move the images' master branch to the first commit",
      Task: MoveBranchTask,
      followUp: (
        <>
          <p>
            Not only was the state of the file in the montage repo restored to
            the original version of the file, but all the intermediate results
            were recreated, as well. Via the Pachyderm API and the pachctl tool,
            you reproduce any results ever created.
          </p>
          <p>
            Making reproducibility easy-to-use is what makes Pachyderm
            cost-effective for your MLOps.
          </p>
        </>
      ),
    },
    {
      header: "Confirm that the montage's original version is restored",
      info: (
        <>
          <p>
            The job triggered by repointing the master branch should take very
            little time to run. When it&apos;s done, you should see that the
            original version of the montage has been restored.
          </p>
        </>
      ),
      taskName: 'Look at the restored montage file',
      Task: CheckMontageVersionTask,
    },
  ],
};

export default DataLineageReproducibility;
