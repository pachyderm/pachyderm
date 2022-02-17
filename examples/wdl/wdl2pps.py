import argparse
import json

import python_pachyderm
import WDL

doc: WDL.Document = None
workflow: WDL.Workflow = None
workflow_variables = {}
workflow_name = ""
tasks = set()


def workflow_name():
    return doc.workflow.name


def tasks():
    return {task.name: task for task in doc.tasks}


def get_task_from_call(call: WDL.Tree.Call) -> WDL.Tree.Task:
    return tasks()[call.name]


def create_repo(repo_name, glob="/"):
    return {"pfs": {"repo": repo_name, "glob": glob}}


def eval_bindings(call: WDL.Tree.Call):
    # evaluate variables and values and create bindings
    bindings = WDL.Env.Bindings()
    task = get_task_from_call(call)

    call_variables = {}
    for (
        input_arg,
        input_value_ref,
    ) in call.inputs.items():  # call.inputs is dict(str -> expr)
        input_arg, input_value_ref = str(input_arg), str(input_value_ref)
        if input_value_ref in workflow_variables:
            call_variables[input_arg] = workflow_variables[input_value_ref]
    for key, value in call_variables.items():
        bindings = bindings.bind(key, WDL.Value.String(value))

    # evaluate call outputs, so that they become the inputs of the next call
    output_files = []
    for tout in task.outputs:
        value = tout.expr.eval(bindings, WDL.StdLib.Base("1.0")).value
        if isinstance(tout.type, WDL.Type.File):
            output_files.append(value)
            value = f"/pfs/{call.name}/{value}"
        workflow_variables[".".join([call.name, tout.name])] = value

    return bindings


# creates a pach pipeline without inputs
def migrate_call_to_spec(call: WDL.Tree.Call):
    task = tasks()[call.name]
    bindings = eval_bindings(call)
    # assume there is only one dependency
    dependency = call.workflow_node_dependencies.pop()
    if dependency.startswith("gather-call-"):
        # gather-call-x
        dependency = dependency.split("-")[2]
    elif dependency.startswith("decl-") or dependency.startswith("call-"):
        dependency = dependency.split("-")[1]

    spec = {
        "pipeline": {"name": task.name},
        "transform": {
            "image": task.runtime["docker"]
            .eval(bindings, WDL.StdLib.Base("1.0"))
            .value,
            "cmd": ["/bin/bash"],
            "stdin": [],
        },
        "resource_requests": {},
    }
    if "cpu" in task.runtime:
        spec["resource_requests"]["cpu"] = (
            task.runtime["cpu"].eval(bindings, WDL.StdLib.Base("1.0")).value
        )
    if "memory" in task.runtime:
        spec["resource_requests"]["memory"] = str(
            task.runtime["memory"].eval(bindings, WDL.StdLib.Base("1.0")).value
        )
    stdin = spec["transform"]["stdin"]
    stdin.append("set -vx")
    stdin.append("cd /pfs/out")
    stdin.append(f"{task.inputs[0].name}=$(find /pfs/{dependency}/ -type f)")
    for cmd in (
        task.command.eval(bindings, WDL.StdLib.Base("1.0")).value.strip().split("\n")
    ):
        cmd = cmd.strip()
        stdin.append(cmd)

    spec["input"] = create_repo(dependency, "/")

    return spec


def migrate_scatter_to_spec(scatter: WDL.Tree.Scatter):
    # assume scatters are 1 level deep, and consists of only one call
    call = [x for x in scatter.body if isinstance(x, WDL.Tree.Call)][0]
    task = get_task_from_call(call)

    # assume there is only one dependency
    dependency = scatter.workflow_node_dependencies.pop()
    dependency_name = dependency.split("-")[1]

    # TODO evaluate variables and values and create bindings
    bindings = eval_bindings(call)
    spec = {
        "pipeline": {"name": call.name},
        "transform": {
            "image": task.runtime["docker"]
            .eval(bindings, WDL.StdLib.Base("1.0"))
            .value,
            "cmd": ["/bin/bash"],
            "stdin": [],
        },
        "resource_requests": {},
    }
    if "cpu" in task.runtime:
        spec["resource_requests"]["cpu"] = (
            task.runtime["cpu"].eval(bindings, WDL.StdLib.Base("1.0")).value
        )
    if "memory" in task.runtime:
        spec["resource_requests"]["memory"] = (
            task.runtime["memory"].eval(bindings, WDL.StdLib.Base("1.0")).value
        )
    stdin = spec["transform"]["stdin"]
    stdin.append("set -vx")
    stdin.append("mkdir -p /pfs/out/${PACH_DATUM_ID}")
    stdin.append("cd /pfs/out/${PACH_DATUM_ID}")
    stdin.append(f"inputFiles=($(ls -d /pfs/{dependency_name}/**))")
    stdin.append(f"{task.inputs[0].name}=${{inputFiles[0]}}")
    command = []
    for part in task.command.parts:
        if isinstance(part, str):
            command.append(part.strip())
        elif isinstance(part, WDL.Expr.Placeholder):
            command.append(f"${str(part.expr).strip()}")
    stdin.append(" ".join(command))

    spec["input"] = create_repo(dependency_name, "/**")

    return spec


def process_workflow(wdl_file, config_dict=None):
    global doc
    # parse WDL
    doc = WDL.load(wdl_file)

    # workflow_variables keep track of the name and value of variables for the entire workflow
    for k, v in config_dict.items():
        k = k.split(".")[1] if "." in k else k
        workflow_variables[k] = v

    # compile workflow body to pipeline specs
    pipelines = []
    for n in doc.workflow.body:
        if isinstance(n, WDL.Tree.Call):
            pipelines.append(migrate_call_to_spec(n))
        elif isinstance(n, WDL.Tree.Scatter):
            pipelines.append(migrate_scatter_to_spec(n))

    return pipelines


def run_pipelines(pipelines):
    # Run pipelines in Pachyderm
    # need calls for topological ordering
    client = python_pachyderm.Client()
    for p in pipelines:
        p["update"] = True
        p["reprocess_spec"] = "every_job"
        req = python_pachyderm.parse_dict_pipeline_spec(p)
        client.create_pipeline_from_request(req)


def main():
    parser = argparse.ArgumentParser(description="Pass WDL and config file.")
    parser.add_argument("--wdl", type=str, help="The WDL workflow file.")
    parser.add_argument(
        "--config", type=str, help="A file with input key, value pairs for workflow."
    )
    parser.add_argument("--dry-run", action="store_true", default=False)

    args = parser.parse_args()

    config_dict = json.load(open(args.config, "r")) if args.config else {}

    pipelines = process_workflow(args.wdl, config_dict)
    if args.dry_run:
        print(json.dumps(pipelines))
    else:
        print("Running pipelines for realz!")
        run_pipelines(pipelines)


if __name__ == "__main__":
    main()
