package v2_8_0

func ppsCollections() []*postgresCollection {
	return []*postgresCollection{
		newPostgresCollection("project_defaults"),
	}
}
