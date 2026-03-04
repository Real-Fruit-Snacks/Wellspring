package loader

import "sync"

var (
	mu      sync.RWMutex
	loaders = make(map[string]Loader)
)

func Register(l Loader) {
	mu.Lock()
	defer mu.Unlock()
	loaders[l.Name()] = l
}

func Get(name string) (Loader, bool) {
	mu.RLock()
	defer mu.RUnlock()
	l, ok := loaders[name]
	return l, ok
}

func List() []Loader {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Loader, 0, len(loaders))
	for _, l := range loaders {
		result = append(result, l)
	}
	return result
}

func ListSupported(os, arch string) []Loader {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Loader, 0)
	for _, l := range loaders {
		if l.Supports(os, arch) {
			result = append(result, l)
		}
	}
	return result
}
