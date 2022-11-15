import React from 'react';

import {Story} from '@pachyderm/components';

import styles from './IncrementalScalability.module.css';
import AddImagesTask from './tasks/AddImagesTask';
import MinimizeTask from './tasks/MinimizeTask';
import {ReactComponent as DatumVisualization} from './Visualization.svg';

const IncrementalScalability: Story = {
  name: 'Incremental Scalability Saves You Time and Money',
  sections: [
    {
      header: 'About Incremental Scalability',
      info: (
        <p>
          Pachyderm&apos;s data-driven pipelines and reproducibility combine to
          provide incremental scalability, which can save you time and money in
          development and MLOps. As shown in the last unit, reproducibility is
          easy-to-use in Pachyderm. Pachyderm&apos;s immutable commit
          architecture makes reproducibility cost-effective, too. Since
          Pachyderm stores all results in immutable commits, only new results
          need to be computed. Any results that rely on old data can be
          reproduced. We call that incremental scalability: only scaling as
          needed to process new data.
        </p>
      ),
    },
    {
      header: 'Add files to the images repo ',
      taskName: 'Add files to the images repo',
      info: (
        <p>
          We&apos;ll add more files to the images repo so that we can see only
          the added images are processed.
        </p>
      ),
      Task: AddImagesTask,
    },
    {
      header: 'Examine skipped and processed data stats',
      taskName: 'Confirm skipped datums',
      info: (
        <>
          <p>
            After minimizing this overlay, navigate to the edges pipeline and
            find the topmost commit in the master branch. That commit contains
            the files you just added, along with the files you added in the
            previous steps.
          </p>
          <p>
            Click on the <q>Linked Job</q> for that top commit. Scroll down
            until you see the Datums section for that job. The datums processed
            attribute is the number of files actively processed, datums skipped
            shows the number of files that were skipped, and datums total is the
            number of files total in the commit, for this pipeline.
          </p>
        </>
      ),
      Task: MinimizeTask,
    },
    {
      header: 'Pachyderm saves you time and money',
      info: (
        <>
          <p>
            Pachyderm&apos;s incremental scalability makes reproducibility fast,
            efficient and cost-effective, saving you time in development and
            money in operations.
          </p>
          <p>
            If Pachyderm wasn&apos;t keeping track of all your data and code,
            you might have to reprocess everything. As this graph shows, using
            Pachyderm turns rapidly-increasing cost nightmare into a flat,
            constant, predictable spend.
          </p>
          <div className={styles.svgWrapper}>
            <DatumVisualization />
          </div>
        </>
      ),
      followUp: (
        <>
          <p>
            Previously in this tutorial, you reset the head commit in the master
            branch to an existing commit to reproduce an output throughout the
            DAG, examining the reproduction of results in the montage repo. When
            you performed that task, a job was created which actively processed
            zero datums. You can look at that job and confirm it the same way
            you confirmed that only new files were processed here.
          </p>
          <p>
            Of course, you can change these default behaviors. It is possible to
            configure a pipeline to always process all data, regardless of
            previous processing status. You can also configure a pipeline to
            have dynamic scalability, and process data in parallel.
          </p>
          <p>
            Pachyderm is the data foundation for efficient MLOps. You can
            further explore its capabilities via these examples or Pachyderm
            Notebooks
          </p>
        </>
      ),
    },
  ],
};

export default IncrementalScalability;
