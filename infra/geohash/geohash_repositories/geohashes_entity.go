package geohash_repositories

type Geohash struct {
	Geohash string `gorm:"type:text;primaryKey"`
	ISO3    string `gorm:"type:varchar(3);index"`
} //@name Geohash

func (Geohash) TableName() string {
	return "geohashes"
}
