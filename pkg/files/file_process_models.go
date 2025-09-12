package files

import "io"

type ImageType string

const (
	ImageTypeProfile   ImageType = "profiles"
	ImageTypeEvents    ImageType = "events"
	ImageTypeLocations ImageType = "locations"
	ImageTypeFeedBlock ImageType = "feed_blocks"
) //@Field ImageType

func (e ImageType) String() string {
	return string(e)
}

type ImageContents struct {
	Name      string    `json:"name" binding:"required"`
	URL       string    ``
	ImageType ImageType `json:"image_type" binding:"required"`
	Content   io.Reader `json:"content" binding:"required"`
} //@Field ImageContents

// ImagesResizingWithPrefixes mapping with images prefix_<image name>
var ImagesResizingWithPrefixes = map[string]int{
	"thumbnail_": 200,
	"sm_":        350,
	"md_":        800,
	"lg_":        1024,
}
