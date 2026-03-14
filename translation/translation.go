package translation

import (
	"context"

	"golang.org/x/text/language"
)

type Translator interface {
	Translate(ctx context.Context, from *language.Tag, text string) (TranslatedText, error)
	TranslateTo(
		ctx context.Context,
		from *language.Tag,
		to language.Tag,
		text string,
	) (TranslatedText, error)
}

type TranslatedText struct {
	SourceText     string       `json:"sourceText"`
	TranslatedText string       `json:"translatedText"`
	FromLanguage   language.Tag `json:"fromLanguage"`
	ToLanguage     language.Tag `json:"toLanguage"`
}
