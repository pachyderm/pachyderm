# Dev 1 Stack

> Used for testing pulumi infra changes for the pulumi core project.

# Overview

# Required Environment Variables

> pulumi config set --secret rdsPGDBPassword $IAC_CI_DB_PASSWORD

# Tags

Tag all resources so we can track billing and usage.

```
		Tags: pulumi.StringMap{
			"Project": pulumi.String("Feature Testing"),
			"Service": pulumi.String("CI"),
			"Owner":   pulumi.String("pachyderm-ci"),
			"Team":    pulumi.String("Core"),
		},
```