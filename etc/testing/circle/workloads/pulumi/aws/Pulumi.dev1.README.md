# Dev 1 Stack

> Used for testing pulumi infra changes for the pulumi core project.

# Overview

This stack is reserved for testing pulumi IAC itself. The `pulumi up` is usually executed locally not from CI. This provides an environment to test IAC changes before they are commited and used by the other stacks. 

> ITS A GOOD IDEA TO ASK IF ANYONE ELSE IS USING THIS STACK. You can always create a DEV2 etc if needed.

# Environment Variables

The test DB test password is reqiored and needs to be set like so. Never check these values in. CI Uses its own stored in CircleCI.

> pulumi config set --secret rdsPGDBPassword $IAC_CI_DB_PASSWORD

`$ENT_ACT_CODE` - Test enterprise env variable key needs to be set locally from where ever the `pulumi up` is called from. 

The following can be overwritten 

`pulumi config set pachdVersion << pachdVersion >>`

`pulumi config set helmChartVersion << helmChartVersion >>`

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