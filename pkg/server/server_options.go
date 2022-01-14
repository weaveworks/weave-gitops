package server

// ApplicationsOptions includes all the options that can be set for an
// ApplicationsServer.
type ApplicationsOptions struct {
	ClientGetter ClientGetter
}

// ApplicationsOption defines the signature of a function that can be used
// to set an option for an ApplicationsServer.
type ApplicationsOption func(*ApplicationsOptions)

// WithClientGetter allows for setting a ClientGetter.
func WithClientGetter(clientGetter ClientGetter) ApplicationsOption {
	return func(args *ApplicationsOptions) {
		args.ClientGetter = clientGetter
	}
}
