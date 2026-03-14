package translation

import (
	"maps"
	"slices"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

var (
	mu        sync.RWMutex
	languages = make(map[language.Tag]Descriptor)
)

func Register(descriptors ...Descriptor) {
	mu.Lock()
	defer mu.Unlock()

	for _, desc := range descriptors {
		languages[desc.Tag] = desc
	}
}

func GetAvailableLanguages() []Descriptor {
	mu.RLock()
	defer mu.RUnlock()

	ls := slices.Collect(maps.Values(languages))
	slices.SortFunc(ls, func(a, b Descriptor) int {
		return strings.Compare(a.Name, b.Name)
	})

	return ls
}

func GetDescriptorByTag(tag language.Tag) (Descriptor, bool) {
	ld, ok := languages[tag]

	return ld, ok
}
