package install

// InstallOptions carries whole-install operation configuration.
type InstallOptions struct {
	WithOptional bool
}

func DefaultInstallOptions() InstallOptions {
	return InstallOptions{}
}
