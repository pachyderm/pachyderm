# Performance Testing
- This folder contains everything needed to do performance testing against a Console instance
- The performance tests utilize [Playwright](https://playwright.dev/) to define and run simultaneous user journeys
- The [cluster-setup](cluster-setup/README.md) folder allows you to quicky create many projects, repos, pipelines, jobs, and files
- The [load-tests](load-tests/README.md) folder allows you to run user journeys and report the results
- A performance cluster is already set up at [http://104.154.210.180/](http://104.154.210.180/)

## Example test performance scenario
- You have a new feature and you think it might change Console's performance
- Start by spinning up the [performance cluster](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#bbfad0e8eeda47989e5e02491a33ccf9) and [SQL instance](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#e6806c07c0ee4a3ea8500ac048b54567)
- [Deploy](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#716259dc5ae24242913fc40dd84beba6) the master tag to the performance cluster
- Run the [load tests](load-tests/README.md) against master and note any failures
- Interact with the site while the load tests are running and note any particular slowness
- [Deploy](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#716259dc5ae24242913fc40dd84beba6) your branch tag to the performance cluster
- Run the [load tests](load-tests/README.md) against your branch and note any differences against master
- Once again, interact with the site while the load tests are running for comparison
- After you are finished, spin down the [performance cluster](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#bbfad0e8eeda47989e5e02491a33ccf9) and [SQL instance](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#e6806c07c0ee4a3ea8500ac048b54567)

## Managing GCP Instances
See [Mangaging GCP Instances](https://www.notion.so/Managing-GCP-Instances-2aeb7aeb36c14598a0130356cb755f2c#4c34936462fe41fba80106cbe35fae90)

## Setting up a performance cluster
See [cluster-setup](cluster-setup/README.md)

## Running load-tests against a cluster
See [load-tests](load-tests/README.md)

## Performance cluster location
[http://104.154.210.180/](http://104.154.210.180/)
