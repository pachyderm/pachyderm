
1. we want to version the generated golang test cases that include the resources embedded in the golang code (what the hack/gen-test-target.sh is doing)
2. the generated code is a known, passing test case that is used to compare with PR changes.
3. if the author of the PR is making changes- they should regen the test case
4. we should also do a `kustomize build | kubectl apply --validate --dry-run -f -` so kubectl can check on the indentation in the yaml, bogus values, schema checks. It will validate syntax and parameters and values.


5. gotest close to package (manifests)
6. golang is not brittle


