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

Currently, protobuf code is generated in a separate node package [@pachyderm/proto](https://www.npmjs.com/package/@pachyderm/proto?activeTab=dependencies). To update the version of the protos update the dependency in `package.json`.
```bash
npm install @pachyderm/proto@{DESIRED_VERSION} --save-exact
```

Update `version.json` to reference the version of Pachyderm you want to pull. This should match the version of the protos being used in the `package.json`.

## Testing

#### Deploy pachyderm locally
1. [Install helm](https://helm.sh/docs/intro/install/).
2. Grab the latest pachyderm helm chart: `helm repo add pachyderm https://pachyderm.github.io/helmchart`
3. If you haven't already, start up minikube
4. Install pachyderm locally: `helm install pachyderm --set deployTarget=LOCAL --version {DESIRED_VERSION} pachyderm/pachyderm`
5. Delete your existing pre-2.0 pachctl config file: `rm ~/.pachyderm/config.json`
6. Run `pachctl port-forward`

After deploying pachyderm locally you can run all tests with the following command
```bash
npm run test
```

As of now we are testing the builders and service functions. 
### Builder 
These tests do not require you to deploy pachyderm locally to run against. Tests should be in the `__tests__` directory inside the `builders` folder. For each builder function, the default parameters along with parameters that override the default values should be tested.
### Service
Tests should be in the `__tests__` directory inside the `services` folder. These tests should use the `pachClient` to hit pachyderm and test that the expected behavior of each service call is satisfied.

## Contributing to the library

1. Code your change on a feature branch
2. Add the appropriate tests and make sure existing tests pass
3. Create a PR

Once your PR is approved, it will be merged into the library.
