# Development Guide

1. [Coding Standards](https://github.com/pachyderm/company/blob/master/handbook/frontend.md)
1. [Writing PRs](#writing-prs)
1. [Working with GraphQL](#working-with-graphql)
1. [Working with the Components Library](#working-with-the-component-library)
1. [Working with the node Pachyderm client](#working-with-the-node-pachyderm-client)

## Writing PRs

Your PRs should be concise and descriptive to the contents of your changeset. If following a Jira ticket, your branch should be named after the ticket like `FRON-100`. Our current process is to wait for at least **1** reviewer before merging the PR. Try to break down larger PRs into digestible chunks, if it can be separated into smaller changes. Try to exclude large procedural changes like moving a directory or applying linter changes from new features, and instead publish those changes as clearly marked, independent PRs.

### Do

- Try to provide a brief description of what changes you're bringing in, and why, if necessary.
- Provide screenshots as an easy glimpse into the contents of the PR if there are significant visual components, such as bringing in new UI elements.
- Include any additional details required to be able to see and run the changeset. E.g. Any preliminary setup steps, necessary configurations, or helpful tips.
- Include any details about changes external to the PR. E.g. A link to changes in CI, an example of a bot in action, or a link to a cloud console.

## Working with the Component Library

The components under `frontend/components` are remnants of a shared components library between different pachyderm projects, usually corresponding to a component in the [Design System](https://www.figma.com/file/AAydA6OQhML401OgwNLesv/Design-System?node-id=0%3A1&t=MbAAIPq0HDW9BDXr-0). Updates to the components library get automatically published to our [Storybook](https://console-components.netlify.app/).

### To run storybook locally

```bash
cd frontend
npm run storybook
```

### How to add a new component

1. Include a storybook file
2. Add any tests, if necessary
3. Add new components to `index.ts`

### How to add icons

We typically export icons provided to us from the design team on Figma. The final SVG file should have one `svg` parent and one path child with no ids set, and `fill="currentcolor"` on the parent.

1. Install this SVG export extension for Figma <https://www.figma.com/community/plugin/814345141907543603/SVG-Export>
2. Set the default options "Use currentcolor as fill" and "Remove all fills" to true
3. Export your icons and add them under the `SVG` component in this repo
4. Update `index.ts` and `Svg.stories.tsx` as appropriate.

## Working with the node Pachyderm client

### Updating protobuf code

1. Install [jq](https://stedolan.github.io/jq/download/)
1. Update the pachyderm version in `version.json`
1. Change directories to `backend/src/proto`
1. Install with `npm i`
1. Generate new protos with `npm run build:proto`

**_Note: You may want to delete your `node_modules` folder generated in this step after building new protos_**
