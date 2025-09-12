package geometry

import "github.com/mmcloughlin/geohash"

type Geohash struct {
	value string

	precision uint // default 6
	lat       float64
	lon       float64
}

func NewGeohash() *Geohash {
	return &Geohash{}
}

func (g Geohash) String() string {
	return g.value
}

func (g Geohash) Hash() string {
	return g.value
}

func (g Geohash) Precision() uint {
	return g.precision
}

func (g Geohash) LatLon() (float64, float64) {
	return g.lat, g.lon
}

func (g *Geohash) FromHash(hash string) *Geohash {
	g.value = hash
	g.precision = uint(len(hash))
	g.lat, g.lon = geohash.Decode(hash)

	return g
}

func (g *Geohash) FromLatLon(lat float64, lon float64, precision ...uint) *Geohash {
	var precisionValue uint = 6
	if len(precision) > 0 {
		precisionValue = precision[0]
	}

	g.value = geohashFromLatLon(lat, lon, precisionValue)
	g.lat = lat
	g.lon = lon
	return g
}

func (g *Geohash) From(geohash Geohash) *Geohash {
	g.value = geohash.value
	g.precision = geohash.precision
	g.lat = geohash.lat
	g.lon = geohash.lon

	return g
}

func (g Geohash) PrecisionVariations() []Geohash {
	var geohashes []Geohash
	for _, hash := range geohashPrecisionVariations(g.value) {
		var g = NewGeohash().FromHash(hash)
		geohashes = append(geohashes, *g)
	}

	return geohashes
}

func geohashFromLatLon(lat float64, lon float64, precision uint) string {
	return geohash.EncodeWithPrecision(lat, lon, precision)
}

func geohashPrecisionVariations(geohash string) []string {
	geohashes := make([]string, 0)
	for geohashLength := len(geohash); geohashLength > 1; geohashLength-- {
		geohashes = append(geohashes, geohash[0:geohashLength])
	}

	return geohashes
}
