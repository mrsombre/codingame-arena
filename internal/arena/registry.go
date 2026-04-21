package arena

import "sort"

var factories = make(map[string]GameFactory)

// Register adds a GameFactory to the global registry. Called by engine packages in init().
func Register(f GameFactory) {
	factories[f.Name()] = f
}

// GetFactory returns the factory registered under name, or nil.
func GetFactory(name string) GameFactory {
	return factories[name]
}

// Games returns the sorted list of registered game names.
func Games() []string {
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
