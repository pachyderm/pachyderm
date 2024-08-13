# Bazel + Console

### External Dependencies
If you are on an M-series Mac, you will need to install some external dependencies,
as the canvas module does not distribute ARM builds and therefore needs to be built
from source. See the [README.md](./README.md) for more details.

### PNPM
Prior to the introduction of the bazel tooling to this project, all developers 
used `npm` for project and dependency management. The
[JavaScript Rules developed by Aspect](https://docs.aspect.build/rules)
opt, instead, to use `pnpm` for dependency management. These tools are mostly
compatible but it is worth noting this change. `pnpm` generates and manages
`pnpm-lock.yaml` files as opposed to the `npm` managed `package-lock.json`
files. Additionally, the `node_modules/` directories will have a different
structure when built using `pnpm`. 

### Node Modules
The following command generates the `node_modules/` directory using bazel:
```
bazel run //:pnpm -- --dir $PWD install
```
when run within the `backend/` and `frontend/` subdirectories.

Once installed you can continue to run all commands through bazel using the
format:
```
bazel run //:pnpm -- --dir $PWD <command>
```
or you can use `npm`, if installed.
