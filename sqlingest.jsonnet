local pachyderm = {
	local newPipeline(name, input, transform) = {
		pipeline: {
				name: name,
		},
		transform: transform,
		input: input,	
	},
	local pachtf(args, env={}) = {
		image: "pachyderm/pachtf:latest",
		env: env,
		cmd: ["/app/pachtf"] + args,
	},

	// TODO: secret_name should specify a kubernetes secret
	local sqlIngestCron(name, url, query, format, cronSpec, secret_name) =
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
	],
	sqlIngestCron :: sqlIngestCron,
};

// Above here are Pachyderm provided functions.
// Below here is user code
pachyderm.sqlIngestCron(
  name="ingest",
  url="mysql://root@mysql:3306/test_db",
  query="SELECT * FROM test_data",
  format="csv",
  cronSpec="@every 10s",
  secret_name="", // TODO: k8s secret
)
