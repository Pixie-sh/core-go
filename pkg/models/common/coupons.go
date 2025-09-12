package common

import (
	"github.com/pixie-sh/core-go/pkg/models"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "Active"
	CouponStatusInactive CouponStatus = "Inactive"
) //@Field CouponStatus

var CouponStatusList = []CouponStatus{
	CouponStatusActive,
	CouponStatusInactive,
}

type Coupon struct {
	models.SoftDeletable

	ID                *uid.UID        `json:"id,omitempty"`
	Name              string          `json:"name,omitempty"`
	Code              string          `json:"code"`
	Status            CouponStatus    `json:"status,omitempty"`
	Discount          uint64          `json:"discount"`
	StripeCouponID    string          `json:"stripe_coupon_id,omitempty"`
	StripePromotionID string          `json:"stripe_promotion_id,omitempty"`
	UsageHistory      []CouponHistory `json:"usage_history,omitempty"`
}

type CouponHistory struct {
	models.SoftDeletable

	ID        uid.UID `json:"id"`
	CouponID  uid.UID `json:"coupon_id"`
	PartnerID uint64  `json:"partner_id"`
}

type CouponValidityResponse struct {
	Valid  bool    `json:"valid"`
	Coupon *Coupon `json:"coupon" json:"coupon,omitempty"`
}

type CreateCouponRequest struct {
	Name     string       `json:"name" validate:"required"`
	Code     string       `json:"code" validate:"required"`
	Status   CouponStatus `json:"status" validate:"required,isEnum:CouponStatus"`
	Discount uint64       `json:"discount" validate:"required"`
}

type UpdateCouponRequest struct {
	Name   string       `json:"name" validate:"required"`
	Status CouponStatus `json:"status" validate:"required,isEnum:CouponStatus"`
}
