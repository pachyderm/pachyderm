---
# YAML header
ignore_macros: true
---

<!-- git-snippet: enable -->
# Pipeline Templates

Pachyderm pipeline's specifications are intuitive, simple, and language agnostic.
They are, however, very static.

A **pipeline template** is a thin wrapping layer atop of your JSON file, 
allowing you to **parameterize a pipeline specification file**, 
thus adding a dynamic component to the creation and update of pipelines.

With pipeline templates, you can easily reuse the baseline of a given pipeline
while experimenting with various values of specific fields.

## Template

Pachyderm's pipeline templates are written in 
the open-source templating language [jsonnet](https://jsonnet.org/){target=_blank}.
A template wraps the baseline of a JSON file into a function, 
allowing the injection of parameters to a pipeline specification file. 

All pipeline templates are `.jsonnet` files.

As an example, check the file `edges.jsonnet` below. It is a "templated" version
of the edges pipeline in the opencv example, parameterized to inject a name modifier 
and an input repository name into the pipeline specifications.

Note that the parameter `src` sits in place of what would have been
the value of the `repo` field, 
as a placeholder for any parameter that will be passed to the template.

!!! Example 
    In this snippet of the template, `src` is a parameter, adding the flexibility to pass any input repository, where, in its original form, the pipeline had `images` as input repository.

    ```yaml
    input: {
        pfs: {
          name: "images",
          glob: "/*",
          repo: src,
        }
      },
    ```

See the full `edges.jsonnet` here:
```yaml
{{ gitsnippet('pachyderm/pachyderm', 'examples/opencv/templates/edges.jsonnet', 'master') }}
```

Or check our full ["templated" opencv example](../../../../../examples/opencv/templates/README){target=_blank}.

To instanciate a template:

- add the `--jsonnet` flag to your `pipeline create` or `pipeline update` commands, followed by a local path to your jsonnet file or an url.
- add `--arg <parameter-name>=value` for each variable.

!!! Example
    ```shell
    pachctl create pipeline --jsonnet templates/edges.jsonnet --arg suffix=1 --arg src=images
    ```

    The command above will create a JSON file named `edges-1` and point the input repository to `images`.

!!! Information 
    Read jsonnet's complete [standard library documentation](https://jsonnet.org/ref/stdlib.html){target=_blank} to know all the variables types, string manipulation and mathematical functions, or assertions available to you.


At the minimum, your function should always have a parameter that acts as a name modifier. 
Pachyderm's pipeline names are unique. 
You can quickly generate several pipelines from the same template
by adding a prefix or a suffix to its generic name.

!!! Info 
    Your .jsonnet file can create multiple pipelines at once.

## Use Cases

Using pipeline templates, you could pass different images
to the transform section of an otherwise identical JSON specification file
to train multiple models on the same dataset,
or switch between one input repo holding test data to another holding production data by parameterizing the input repo field. 

During the development phase of a pipeline, 
it can be helpful to pass the tag of an image as a parameter: 
each re-build of the pipeline's code requires you to increment your tag value;
passing it as a parameter will save you the time to update your JSON specifications.
You could also consider preparing a library of ready-made templates for data science teams to instantiate, according to their own set of parameters. 

We will let you imagine more use cases in which those templates can be helpful to you.