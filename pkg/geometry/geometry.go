package geometry

import (
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

const WGS84_SRID = 4326

func CreateWkbPointFromLatLng(lat, lng float64) ([]byte, error) {
	point := geom.NewPointFlat(geom.XY, []float64{lng, lat}) // (lon, lat) order
	point.SetSRID(WGS84_SRID)

	wkbPoint, err := wkb.Marshal(point, wkb.NDR)
	if err != nil {
		return nil, err
	}

	return wkbPoint, nil
}
