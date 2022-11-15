import React from 'react';

import {Story} from '@pachyderm/components';

import MinimizeTask from './tasks/MinimizeTask';

const EasyToUnderstandDataLineage: Story = {
  name: 'Easy-to-understand Data Lineage',
  sections: [
    {
      header: 'Global identifiers tie together commits and code',
      info: (
        <>
          <p>
            Data lineage means you can easily find what version of what data was
            used with what code to produce a particular output.
          </p>
          <p>
            Pachyderm&apos;s versioned data repositories provide a layered
            history of data through commits. Commits in Pachyderm are immutable,
            providing auditable evidence.
          </p>
          <p>
            When data is put into Pachyderm, the commit produces a global
            identifier that easily allows you to link code, data, models and
            parameters together.
          </p>
          <p>
            You can trace that global identifier through the system to find
            anything that went into creating a particular result.
          </p>
        </>
      ),
      taskName: 'Look at the global identifier for a job',
      Task: MinimizeTask,
      followUp:
        'You can also use the Pachyderm tool pachctl to trace an input commit throughout multiple pipelines and DAGs.',
    },
  ],
};

export default EasyToUnderstandDataLineage;
