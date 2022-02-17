# WDL to Pachyderm Pipelines

A Python script that converts Workflow Description Langauge (WDL) to Pachyderm pipelines.

## Requirements

- [`pachctl`](https://docs.pachyderm.com/2.0.x-rc/getting_started/local_installation/#install-pachctl) A CLI tool for managing data in Pachyderm
- `miniwdl` for parsing WDL
- `python-pachyderm` for creating pipelines in Pachyderm
- a working Pachyderm instance

To install all python dependencies run:
```
pip install -r requirements.txt
```

Run
```
python wdl2pps.py --help
```

## Examples

> Note: if you just want to get the pipeline JSON, you can run `python wdl2pps.py --dry-run ...`

### Simple linear pipeline

```
pachctl create repo md5Workflow-input
pachctl put file md5Workflow-input@master -f examples/md5/md5sum.input
python wdl2pps.py --wdl examples/md5/md5.wdl --config examples/md5/config.json --dry-run
```

### Scatter-gather

```
pachctl create repo inputFiles
pachctl put file inputFiles@master -f examples/scatter-gather/md5sum.input -f examples/scatter-gather/md5sum.2.input
python wdl2pps.py --wdl examples/scatter-gather/scatter-gather.wdl --config examples/scatter-gather/config.json --dry-run
```

## Core concepts

WDL to PPS

| WDL                 | Pachyderm                     |
|:-------------------:|:-----------------------------:|
| workflow            | implicit DAG                  |
| call                | pipeline                      |
| call.input          | pipeline input                |
| task.runtime        | transform.image + resources   |
