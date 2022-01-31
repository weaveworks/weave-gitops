package types

func getGitopsLabelMap(appName string) map[string]string {
	labels := map[string]string{
		ManagedByLabel: managedByWeaveGitops,
		CreatedByLabel: createdBySourceController,
	}

	if appName != "" {
		labels[PartOfLabel] = appName
	}

	return labels
}
