package enums

type AIEnhancedType string

const (
	AiGeneratedAiEnhancedType         AIEnhancedType = "ai_generated"
	AiGeneratedModifiedAiEnhancedType AIEnhancedType = "ai_generated_modified"
	HumanGeneratedAiEnhancedType      AIEnhancedType = "human_generated"
) //@Field AIEnhancedType

func (s AIEnhancedType) String() string {
	return string(s)
}

func AiEnhancedTypeList() []AIEnhancedType {
	return []AIEnhancedType{
		AiGeneratedAiEnhancedType,
		AiGeneratedModifiedAiEnhancedType,
		HumanGeneratedAiEnhancedType,
	}
}
