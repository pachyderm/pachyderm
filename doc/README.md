# Pachyderm Documentation

These docs are rendered and searchable in our [Documentation Portal](https://docs.pachyderm.com). Here are a few section links for quick access:

- [Getting Started](https://docs.pachyderm.com/latest/getting_started/) — includes local installation and Beginner Tutorial.
- [Concepts](https://docs.pachyderm.com/latest/concepts/) — provides an overview of the main Pachyderm concepts.
- [How-Tos](https://docs.pachyderm.com/latest/how-tos/) — includes how-tos about that describe how you can load your data into Pachyderm, export it to external systems, split your data, and all the other data operaions and features available in Pachyderm. 
- [Pachctl API Documentation](https://deploy-preview-4312--pachyderm-docs.netlify.com/latest/reference/pachctl/pachctl/) — provides an overview of all Pachyderm CLI tool commands.

Have more questions? — Join our [Slack channel](http://slack.pachyderm.io/).

# Can I Contribute to Docs?

Absolutely! We welcome external contributors!  
Before sending a PR, please read our [Documentation Style Guide](https://docs.pachyderm.com/latest/contributing/docs-style-guide/).


# New to Pachyderm's documentation but want to contribute? Set up your Doc development environment.

Our documentation is built on **mkdocs** and runs on **Netlify**. All content is created in markdowm (.md) language.

## 1- Directory structure
Root directory: ```pachyderm/doc```
We are maintaining several versions of the documentation in:
```shell
    - doc
        |- docs
            |- 1.10.x
            |- 1.11.x
            |- master
```
*Note: the **master** version is the future release.*
Each version comes with its own configuration (navigation) file:

```shell
     - doc
        |- mkdocs-1.10.x.yml
        |- mkdocs-1.11.x.yml
        |- mkdocs-master.yml
```

## 2- Run the Documentation locally
### 2.0- Prerequisite
[Docker install](https://docs.docker.com/get-docker/)

### 2.1- Run a *given version* of the documentation locally with hot-reload
This is especially useful in DEV mode (i.e., you are actively writing or editing content on a given version). 
You do not need to understand the inner workings of mkdocs or Netlify. 
You will be able to edit your version of the documentation and immediately check out the outcome on your local browser. 

Follow those 4 steps:

1. Choose the .yml of the doc version you want to consider (Default to mkdocs-1.11.x.yml) and update `doc/docker-compose.yml` accordingly (in 'command' section).

2. Build your Docker image (~10 min). In the `doc` directory, run:
    ```shell
        DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose build mkdocs-serve
    ```
3. Then, run: 
    ```shell
        docker-compose up mkdocs-serve
    ```
4. Check your local server in your preferred browser.
    ```shell
        0.0.0.0:8889
    ```
Your source code is now visible from the container allowing a hot reload.

### 2.2- Run the full doc website (all versions)
Especially useful if you want to have a last look at your final product before committing your work. It will be as close as the production as can be.
1. Build your Docker image (~10 min). In the `doc` directory, run:
    ```shell
        DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose build netlify-dev
    ```
2. In the doc directory, run: 
    ```shell
        docker-compose up netlify-dev
    ```
3. Check your local server in your preferred browser.
    ```shell
        0.0.0.0:8888
    ```
