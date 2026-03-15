package gcloudtranslate

import (
	"context"
	"errors"
	"fmt"

	translate "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/holyheld/gointernals/translation"
	"golang.org/x/text/language"
)

type Translator struct {
	defaultToLang *language.Tag
	client        *translate.TranslationClient
	projectID     string
	parent        string
}

var _ translation.Translator = (*Translator)(nil)

type Option func(*Translator)

func WithDefaultToLang(lang language.Tag) Option {
	return func(t *Translator) {
		t.defaultToLang = &lang
	}
}

func WithClient(cli *translate.TranslationClient) Option {
	return func(t *Translator) {
		t.client = cli
	}
}

const locationIDGlobal = "global"

func getParent(projectID string, location string) string {
	return "projects/" + projectID + "/locations/" + location
}

func WithLocationParent(locationID string) Option {
	return func(t *Translator) {
		t.parent = getParent(t.projectID, locationID)
	}
}

func NewModule(
	ctx context.Context,
	projectID string,
	opts ...func(*Translator),
) (*Translator, error) {
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	t := &Translator{
		client:    client,
		projectID: projectID,
		parent:    getParent(projectID, locationIDGlobal),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t, nil
}

func (t *Translator) Close() error {
	return t.client.Close()
}

func (t *Translator) Translate(
	ctx context.Context,
	from *language.Tag,
	text string,
) (translation.TranslatedText, error) {
	if t.defaultToLang == nil {
		return translation.TranslatedText{}, errors.New(
			"to language is not specified: set defaultToLang or use TranslateTo()",
		)
	}

	res, err := t.translate(ctx, from, *t.defaultToLang, []string{text})
	if err != nil {
		return translation.TranslatedText{}, err
	}

	if len(res) == 0 {
		return translation.TranslatedText{}, errors.New("translate returned no results")
	}

	return res[0], nil
}

func (t *Translator) TranslateTo(
	ctx context.Context,
	from *language.Tag,
	to language.Tag,
	text string,
) (translation.TranslatedText, error) {
	res, err := t.translate(ctx, from, to, []string{text})
	if err != nil {
		return translation.TranslatedText{}, err
	}

	if len(res) == 0 {
		return translation.TranslatedText{}, errors.New("translate returned no results")
	}

	return res[0], nil
}

func (t *Translator) translate(
	ctx context.Context,
	from *language.Tag,
	to language.Tag,
	texts []string,
) ([]translation.TranslatedText, error) {
	options := &translatepb.TranslateTextRequest{
		Contents:           texts,
		TargetLanguageCode: to.String(),
		Parent:             t.parent,
		MimeType:           "text/plain",
	}

	var fromLanguage language.Tag

	if from != nil {
		options.SourceLanguageCode = from.String()
		fromLanguage = *from
	}

	resp, err := t.client.TranslateText(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to translate text: %w", err)
	}

	ret := make([]translation.TranslatedText, len(resp.Translations))
	for i, t := range resp.Translations {
		if i >= len(texts) {
			break
		}

		fromLanguage := fromLanguage

		if t.DetectedLanguageCode != "" {
			parsedTag, err := language.Parse(t.DetectedLanguageCode)
			if err == nil {
				fromLanguage = parsedTag
			}
		}

		ret[i] = translation.TranslatedText{
			//nolint:gosec
			// According to translatepb.TranslateTextResponse.Translations description,
			// "This field has the same length as [`contents`]"
			SourceText:     texts[i],
			TranslatedText: t.TranslatedText,
			FromLanguage:   fromLanguage,
			ToLanguage:     to,
		}
	}

	return ret, nil
}
