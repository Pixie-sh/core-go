package common

type EntityTypeEnum string

const (
	PartnerEntityType EntityTypeEnum = "partner"
	CreatorEntityType EntityTypeEnum = "creator"
)

func (e EntityTypeEnum) String() string {
	return string(e)
}
