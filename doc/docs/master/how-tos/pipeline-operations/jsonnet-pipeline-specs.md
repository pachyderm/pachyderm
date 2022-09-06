---
# YAML header
ignore_macros: true
---

<!-- git-snippet: enable -->
# Jsonnet Pipeline Specifications

!!! Warning
    `Jsonnet pipeline specifications` is an [experimental feature](../../../reference/supported-releases/#experimental).

Pachyderm [pipeline's specification](../../../reference/pipeline-spec){target=_blank} files are intuitive, simple, and language agnostic.
They are, however, very static.

A **jsonnet pipeline specification file** is a thin wrapping layer atop of your JSON file, 
allowing you to **parameterize a pipeline specification file**, 
thus adding a dynamic component to the creation and update of pipelines.

With jsonnet pipeline specs, you can easily reuse the baseline of a given pipeline spec
while experimenting with various values of given fields.

## Jsonnet Specs

Pachyderm's Jsonnet pipeline specs are written in 
the open-source templating language [jsonnet](https://jsonnet.org/){target=_blank}.
Jsonnet wraps the baseline of a JSON file into a function, 
allowing the injection of parameters to a pipeline specification file. 

All jsonnet pipeline specs have a `.jsonnet` extension.

As an example, check the file `edges.jsonnet` below. It is a parameterized version
of the edges pipeline spec `edges.json` in the opencv example, used to inject a name modifier 
and an input repository name into the original pipeline specifications.


!!! Example 
    In this snippet of `edges.jsonnet`, the parameter `src` sits in place of what would have been
    the value of the field `repo`, as a placeholder for any parameter that will be passed to the jsonnet pipeline spec.

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
{{ gitsnippet('pachyderm/pachyderm', 'examples/opencv/jsonnet/edges.jsonnet', '2.3.x') }}
```

Or check our full ["jsonnet-ed" opencv example](https://github.com/pachyderm/pachyderm/tree/2.3.x/examples/opencv/jsonnet){target=_blank}.

To create or update a pipeline using a jsonnet pipeline specification file:

- add the `--jsonnet` flag to your `pipeline create` or `pipeline update` commands, followed by a local path to your jsonnet file or an url.
- add `--arg <parameter-name>=value` for each variable.

!!! Example
    ```shell
    pachctl create pipeline --jsonnet jsonnet/edges.jsonnet --arg suffix=1 --arg src=images
    ```

    The command above will generate a JSON file named `edges-1.json` then create a pipeline of the same name taking the repository `images` as its input.

!!! Information 
    Read jsonnet's complete [standard library documentation](https://jsonnet.org/ref/stdlib.html){target=_blank} to learn about all the variables types, string manipulation and mathematical functions, or assertions available to you.


At the minimum, your function should always have a parameter that acts as a name modifier. 
Pachyderm's pipeline names are unique. 
You can quickly generate several pipelines from the same jsonnet pipeline specification file
by adding a prefix or a suffix to its generic name.

!!! Info 
    Your .jsonnet file can create multiple pipelines at once as illustrated in our [group example](https://github.com/pachyderm/pachyderm/tree/master/examples/group){target=_blank}.

## Use Cases

Using jsonnet pipeline specifications, you could pass different images
to the transform section of an otherwise identical JSON specification file
to train multiple models on the same dataset,
or switch between one input repo holding test data to another holding production data by parameterizing the input repo field. 

During the development phase of a pipeline, 
it can be helpful to pass the tag of an image as a parameter: 
each re-build of the pipeline's code requires you to increment your tag value;
passing it as a parameter will save you the time to update your JSON specifications.
You could also consider preparing a library of ready-made jsonnet pipeline specs for data science teams to instantiate, according to their own set of parameters. 

We will let you imagine more use cases in which those jsonnet specs can be helpful to you.