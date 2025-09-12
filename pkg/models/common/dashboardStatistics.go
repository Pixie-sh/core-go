package common

import "time"

type DashboardStatistics struct {
	VendorCounts  []CountStatistics `json:"vendor_counts,omitempty"`
	CreatorCounts []CountStatistics `json:"creator_counts,omitempty"`
} //@name DashboardStatistics

type CountStatistics struct {
	Date  time.Time `json:"date,omitempty"`
	Count int64     `json:"count,omitempty"`
} //@name CountStatistics
