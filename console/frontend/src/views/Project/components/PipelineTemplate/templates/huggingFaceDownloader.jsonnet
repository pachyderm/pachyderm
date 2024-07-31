/*
title: Hugging Face Downloader
description: "Creates a cron pipeline to download datasets or models from huggingface on demand."
args:
- name: name
  description: The name of the pipeline.
  type: string
- name: spec
  description: The cron spec to use.
  type: string
  default: "@never"
- name: secretName
  description: "The name of the k8s secret containing a huggingface token. The secret itself must contain the token with the key HF_HOME"
  type: string
  default: huggingface-token
- name: type
  description: The type of download (dataset or model).
  type: string
  default: dataset
- name: hf_name
  description: The name of the dataset or model.
  type: string
- name: disable_progress
  description: Whether or not to show progress in the logs. "true" to enable.
  type: string
  default: "false"
*/

local join(a) =
  local notNull(i) = i != null;
  local maybeFlatten(acc, i) = if std.type(i) == "array" then acc + i else acc + [i];
  std.foldl(maybeFlatten, std.filter(notNull, a), []);


local args(hf_name, revision, type) =
  join([
    if revision != "" then ["--revision", revision],
    ["--name", hf_name, "--type", type, "--output_dir", "/pfs/out"],
  ]);


function(name, hf_name, secretName="huggingface-token", revision="", type="dataset", disable_progress="false", allow_patterns="", spec="@never", ignore_patterns="")
  {
    pipeline: { name: name },
    description: "Download HF Dataset: " + name,
    transform: {
      cmd: ["python3", "/app/dataset-downloader.py"] + args(hf_name, revision, type),
      image: "vmtyler/hfdownloader:v0.0.7",
      secrets: [
        {
          name: secretName,
          env_var: "HF_HOME",
          key: "HF_HOME",
        },
      ],
      env: {
        PYTHONUNBUFFERED: "1",
        HF_HUB_DISABLE_PROGRESS_BARS: disable_progress,
      },
    },
    autoscaling: true,
    input: {
      cron: {
        name: "cron",
        spec: spec,
        overwrite: true,
      },
    },
  }
