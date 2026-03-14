package translation

import "golang.org/x/text/language"

type Descriptor struct {
	Tag  language.Tag `json:"tag"`
	Name string       `json:"name"`
}

func NewLanguageDescriptor(lang language.Tag, name string) Descriptor {
	return Descriptor{
		Tag:  lang,
		Name: name,
	}
}
