package translation

import "golang.org/x/text/language"

func BCP47ToISO639(t language.Tag) string {
	base, _ := t.Base()

	return base.ISO3()
}
