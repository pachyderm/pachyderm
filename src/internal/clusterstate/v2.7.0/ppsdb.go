package v2_7_0

func ppsCollections() []*postgresCollection {
	return []*postgresCollection{
		newPostgresCollection("cluster_defaults", nil),
	}
}
