"""Rules for working with Pachyderm."""

PipelineInfo = provider(
    doc = "A Pachyderm pipeline.",
    fields = {
        "name": "The name of the pipeline.",
        "spec": "The pipeline's specification file.",
        "image": "Optionally, an OCI image that is the pipeline's code.",
    },
)

def _pipeline_impl(ctx):
    return [
        PipelineInfo(
            name = ctx.label.name,
            spec = ctx.files.pipeline[0],
            image = ctx.files.image[0],
        ),
    ]

_pipeline = rule(
    implementation = _pipeline_impl,
    attrs = {
        "pipeline": attr.label(
            allow_single_file = True,
            doc = "Pipeline spec file to validate.",
            mandatory = True,
        ),
        "image": attr.label(
            doc = "OCI image that the pipeline runs.",
            mandatory = False,
        ),
    },
)

_SHELL_PREAMBLE = """#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
"""

def _pipeline_shell_command(ctx, name, template):
    pipeline = ctx.attr.pipeline[PipelineInfo]
    template = _SHELL_PREAMBLE + template
    runner = ctx.actions.declare_file(pipeline.name + "_" + name + ".sh")
    ctx.actions.write(
        output = runner,
        content = template.format(
            PACHCTL = ctx.attr._runfiles[0][DefaultInfo].files_to_run.executable.short_path,
            PIPELINE_SPEC = pipeline.spec.short_path,
            IMAGE = pipeline.image.short_path,
        ),
        is_executable = True,
    )
    r = [pipeline.spec, runner]
    if pipeline.image:
        r.append(pipeline.image)
    runfiles = ctx.runfiles(r)
    for r in ctx.attr._runfiles:
        runfiles = runfiles.merge(r.default_runfiles)
    return DefaultInfo(
        executable = runner,
        runfiles = runfiles,
    )

def _pachctl_validate_test_impl(ctx):
    tmpl = """
exec {PACHCTL} validate pipeline --file {PIPELINE_SPEC}
"""
    return _pipeline_shell_command(ctx, "_validate", tmpl)

_pachctl_validate_test = rule(
    implementation = _pachctl_validate_test_impl,
    attrs = {
        "pipeline": attr.label(
            providers = [PipelineInfo],
            doc = "Pipeline spec file to validate.",
        ),
        "_runfiles": attr.label_list(default = ["//src/server/cmd/pachctl"]),
    },
    test = True,
)

def _pachctl_apply_impl(ctx):
    tmpl = """
exec {PACHCTL} update pipeline --file {PIPELINE_SPEC}
"""
    return _pipeline_shell_command(ctx, "_apply", tmpl)

_pachctl_apply = rule(
    implementation = _pachctl_apply_impl,
    attrs = {
        "pipeline": attr.label(
            providers = [PipelineInfo],
            doc = "Pipeline spec file to apply.",
        ),
        "_runfiles": attr.label_list(default = ["//src/server/cmd/pachctl"]),
    },
    executable = True,
)

def _pachctl_push_impl(ctx):
    tmpl = """
find    
exec {PACHCTL} --image={IMAGE} --pipeline={PIPELINE_SPEC}
"""
    return _pipeline_shell_command(ctx, "_update", tmpl)

_pachctl_push = rule(
    implementation = _pachctl_push_impl,
    attrs = {
        "pipeline": attr.label(
            providers = [PipelineInfo],
            doc = "Pipeline spec file to apply.",
        ),
        "_runfiles": attr.label_list(default = ["//src/testing/push-pipeline"]),
    },
    executable = True,
)
    
def pachyderm_pipeline(name, pipeline, image = None, **kwargs):
    _pipeline(
        name = name,
        pipeline = pipeline,
        image = image,
        **kwargs
    )
    _pachctl_validate_test(
        name = name + ".validate",
        pipeline = name,
        **kwargs
    )
    _pachctl_apply(
        name = name + ".apply",
        pipeline = name,
        **kwargs
    )
    _pachctl_push(
        name = name + ".push",
        pipeline = name,
        **kwargs
    )
