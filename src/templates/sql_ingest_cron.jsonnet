local newPipeline(name, input, transform) = {
	pipeline: {
		name: name,
	},
	transform: transform,
	input: input,	
};
local pachtf(args, env={}) = {
	image: "pachyderm/pachtf:latest",
	env: env,
	cmd: ["/app/pachtf"] + args,
};

function (name, url, query, format, cronSpec, secretName)
	local queryPipelineName = name + "_queries";
	[
	newPipeline(
		name=queryPipelineName,
	 	input={
			cron: {
				name: "in",
				spec: cronSpec,
				overwrite: true,
			}
	  },
	  transform=pachtf(["sql-gen-queries", query]),
	),
	newPipeline(
		name=name,
		input={
			pfs: {
				name: "in",
				repo: queryPipelineName,
				glob: "/*",
			},
		},
		transform=pachtf(["sql-ingest", url, format], {
			"PACHYDERM_SQL_PASSWORD": "root",
	  	}),
	)
	]
