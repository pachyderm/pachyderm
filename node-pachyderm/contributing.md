# Contributor guide

## Getting started
Install all dependencies by running
```bash
npm install
```

## Code

### Layout

```
.
├── examples
|   ├── opencv/ - The canonical OpenCV demo rewritten to use node-pachyderm
├── src
│   ├── builders/ - Functions for translating json objects to protobuf types
│   ├── lib/ - Contains types defined by developers
│   ├── services/ - Service calls organized by each proto type
│   └── utils/ - Utility functions
└── version.json - Spec for the version of this library, as well as its pachyderm dependency
```

### Style and Linting

This project uses prettier and eslint to ensure coding consistency. To run the linter locally run
```bash
npm run lint
```
The linter will also run in CI and fail if there are any stylistic discrepancies.

### Updating protobuf code

1. Install [jq](https://stedolan.github.io/jq/download/)
1. Update the pachyderm version in `version.json`
1. Generate new protos with `npm run build:proto`
1. Commit the proto updates and merge into main via pull request

## Testing

#### Deploy Pachyderm locally
1. [Install helm](https://helm.sh/docs/intro/install/)
1. [Install minikube](https://minikube.sigs.k8s.io/docs/start/)
1. Grab the latest pachyderm helm chart: `helm repo add pachyderm https://pachyderm.github.io/helmchart`
1. If you haven't already, start up minikube
1. Install pachyderm locally: `helm install pachyderm --set deployTarget=LOCAL --version {DESIRED_VERSION} pachyderm/pachyderm`
1. Delete any existing pachctl config file: `rm ~/.pachyderm/config.json`
1. Run `pachctl port-forward`

#### Existing Pachyderm deployment
1. WARNING: Keep in mind that the tests will delete any existing repos and pipelines in the activate context.
1. Disable auth for the active context: `pachctl auth deactivate`
1. Run `pachctl port-forward`

After deploying pachyderm locally you can run all tests with the following command
```bash
npm run test
```

As of now we are testing the builders and service functions.
### Builder
These tests do not require you to deploy pachyderm locally to run against. Tests should be in the `__tests__` directory inside the `builders` folder. For each builder function, the default parameters along with parameters that override the default values should be tested.
### Service
Tests should be in the `__tests__` directory inside the `services` folder. These tests should use the `pachClient` to hit pachyderm and test that the expected behavior of each service call is satisfied.

## Examples

### OpenCV
To run the openCV example run the following command
```base
npm run opencv
```
## Contributing to the library

1. Code your change on a feature branch
2. Add the appropriate tests and make sure existing tests pass
3. Create a PR

Once your PR is approved, it will be merged into the library.

## Releases and publishing

1. Update the `version` in `package.json`
1. Update the `version` in `package-lock.json`
1. Commit the version update and merge into main via pull request

A Pachyderm team member will draft a new release on Github and create a tag, for example v0.19.7. CircleCI will automatically publish the updates to npm once a release is created.
