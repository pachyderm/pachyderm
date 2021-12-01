# Node Pachyderm
[![npmversion](https://img.shields.io/npm/v/@pachyderm/node-pachyderm.svg)](https://www.npmjs.com/package/@pachyderm/node-pachyderm)
[![Slack Status](https://badge.slack.pachyderm.io/badge.svg)](http://slack.pachyderm.io)

Official Node Pachyderm client maintained by Pachyderm Inc.
## Installation

```bash
npm i @pachyderm/node-pachyderm
```

## A Small Taste

Here's an example that creates a repo and adds a file:

```javascript
import { pachydermClient } from "@pachyderm/node-pachyderm";

const demo = async () => {
  const pachClient = pachydermClient({
    pachdAddress: "localhost:30650",
  });

  await pachClient.pfs().createRepo({
    repo: { name: "test" },
  });
};

demo();
```
## Contributing
This package is co-maintained by Pachyderm and the community. If you're looking to contribute to the project, this is a fantastic place to get involved. Take a look at [the contributing guide](./contributing.md) for more info (including testing instructions).
