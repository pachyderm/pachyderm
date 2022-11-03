# Pachyderm Component Library

This package contains various components for front end Pachyderm projects. You can find the Storybook docs [here](https://pachyderm-components.netlify.app/).

Our component library is being published to [npm](https://www.npmjs.com/package/@pachyderm/components).

## Local development

If you are using [nvm](https://github.com/nvm-sh/nvm) first run `nvm install`. This will install and switch the version of node to the one defined in the `.nvmrc`. Run `npm install` to install dependencies.

Run the docs locally with: `npm run storybook`
Run the tests with: `npm test`

## Contributing to the library

1. Code your change on a feature branch
2. Add new components to `index.ts`
3. Bump version in `package.json`
4. Run `npm install` to update the version in the package-lock
5. Create a PR

Once your PR is approved, merge your changes and publish the library to npm.
### Publishing to npm

The npm credentials for the our account live in 1password.

MAKE SURE YOU HAVE BUMPED VERSION NUMBER

1. Login to npm: `npm login`
2. Publish to npm: `npm publish`

### Netlify Storybook Docs

Any time you merge something to master and it contains an update to the component library, the docs will be automatically built and publish to netlify.

### Adding Icons
We typically export icons provided to us from the design team on Figma. The final SVG file should have one `svg` parent and one path child with no ids set, and `fill="currentcolor"` on the parent.

1. Install this SVG export extension for Figma https://www.figma.com/community/plugin/814345141907543603/SVG-Export
2. Set the default options "Use currentcolor as fill" and "Remove all fills" to true
3. Export your icons and add them under the `SVG` component in this repo
4. Update `index.ts` and `Svg.stories.tsx` as appropriate. 

