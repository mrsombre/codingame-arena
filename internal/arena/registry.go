package arena

// Factory is the registered GameFactory, set by engine init() via Register.
var Factory GameFactory

// Register sets the global GameFactory. Called by engine packages in init().
func Register(f GameFactory) {
	Factory = f
}
