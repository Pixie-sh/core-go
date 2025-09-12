package language_models

type LanguageEnum string

func (e LanguageEnum) String() string {
	return string(e)
}

const (
	PTLanguageEnum LanguageEnum = "pt_PT"
	ENLanguageEnum LanguageEnum = "en_US"
	NLLanguageEnum LanguageEnum = "nl_NL"
)

var LanguageList = []LanguageEnum{
	PTLanguageEnum,
	ENLanguageEnum,
	NLLanguageEnum,
}
