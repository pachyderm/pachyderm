# Pipeline Templates

Pachyderm pipeline's specifications are intuitive, straightforward, and language agnostic.
However, as easy and lightweight as the JSON format might be, it is very static.
A **pipeline template** is a thin wrapping layer atop your JSON specification file, 
allowing to parameterize a pipeline specification file, thus adding a dynamic component to pipelines.

## Template

Pachyderm's pipeline templates are written in the open-source templating language [jsonnet](https://jsonnet.org/). jsonnet files wrap the baseline of a JSON pipeline into functions, allowing the injection of parameters to a pipeline specifications file. 

To instanciate a template:

- add the `--jsonnet` flag to your `pipeline create` or `pipeline update` commands, followed by a local path to your .jsonnet file or a url
- then `--arg <parameter-name>=value` for each variable.

```shell
pachctl create pipeline --jsonnet templates/edges.jsonnet --arg suffix=1 --arg src=images
```

The command above will create a JSON file named `edges-1` and point the input repository to `images.`


See the file `edges.jsonnet` below, it is a "templated" version of the edges pipeline in the opencv example, parameterized to inject a name modifier and an input repository name into the pipeline specifications.

Note that the parameter name is passed in place of what would have been the value of a JSON field as a placeholder for whatever parameter will be given to the template.

!!! Example "In this snippet of the template, `src` is a parameter, adding the flexibility to pass any input repository, where, in its original form, the pipeline had `images` as input repository."

  ```yaml
  input: {
      pfs: {
        name: "images",
        glob: "/*",
        repo: src,
      }
    },
  ```
All pipeline templates are `.jsonnet` files.

!!! Information 
    Read jsonnet's complete [standard library documentation](https://jsonnet.org/ref/stdlib.html) to know all the variables types, string manipulation, mathematical functions, and assertions available to you.

!!! Example

  ```yaml {{ gitsnippet('pachyderm/pachyderm', 'examples/opencv/templates/edges.jsonnet', '2.1.x') }} ```

At the very minimum, your function should always have a parameter that would act as a name modifier. Pachyderm's pipeline names need to be unique. You can quickly generate several pipelines from the same template by modifying a baseline name and adding a prefix or a suffix to its generic name.

Note that your .jsonnet file can create multiple pipelines at once, possibly a complete DAG.

## Use
Parameterizing pipelines can pass various images to the transform section of your JSON specification file to train multiple models with the same dataset or switch easily between one input repo holding test data to another holding production data. 
During the development phase of a pipeline, it can be helpful to pass the tag of an image as a parameter: each re-build of the pipeline's code requires you to increment your tag value.
You could also consider preparing a library of ready-made templates for data science teams to instantiate, according to their own set of parameters. We will let you imagine the use cases in which those templates can be helpful to you.