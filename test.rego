package hpe.ai.pachyderm.examples.validation

import future.keywords.if

default deny := false

deny if {
    input.pipeline.name == ""
}

deny if {
  input.autoscaling
}

deny if {
  input.description == ""
}

# deny if {
#     not input.pipeline.missing 
# }

