package tags

import "fmt"

func buildTag(scope TagScope, tag string) Tag {
	return Tag(fmt.Sprintf("%s.%s", scope, tag))
}

var (
	TagDealScope = TagScope("deal")

	TagDealNewTag      = buildTag(TagDealScope, "new")
	TagDealTrendingTag = buildTag(TagDealScope, "trending")
)

var (
	TagCreatorScope = TagScope("creator")

	TagCreatorNewTag      = buildTag(TagCreatorScope, "new")
	TagCreatorTrendingTag = buildTag(TagCreatorScope, "trending")
)

var (
	TagPartnerScope = TagScope("partner")

	TagPartnerNewTag      = buildTag(TagPartnerScope, "new")
	TagPartnerTrendingTag = buildTag(TagPartnerScope, "trending")
)
