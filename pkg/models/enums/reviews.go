package enums

type ReviewedByEntityType string

const (
	ReviewedByPartner ReviewedByEntityType = "Partner"
)

func (e ReviewedByEntityType) String() string {
	return string(e)
}

type ReviewedEntityType string

const (
	CreatorReviewed ReviewedEntityType = "Creator"
)

func (e ReviewedEntityType) String() string {
	return string(e)
}

type ReviewStatus string

const (
	PublishedReviewStatus   ReviewStatus = "published"
	PendingReviewStatus     ReviewStatus = "pending"
	ApprovedReviewStatus    ReviewStatus = "approved"
	UnpublishedReviewStatus ReviewStatus = "unpublished"
)

func (e ReviewStatus) String() string {
	return string(e)
}

type ReviewPositiveRemarks string

const (
	RepliesQuicklyReviewPositiveRemarks       ReviewPositiveRemarks = "replies_quickly"
	FollowedInstructionsReviewPositiveRemarks ReviewPositiveRemarks = "followed_instructions"
	CreativityReviewPositiveRemarks           ReviewPositiveRemarks = "creativity"
	AuthenticContentReviewPositiveRemarks     ReviewPositiveRemarks = "authentic_content"
	QualityContentReviewPositiveRemarks       ReviewPositiveRemarks = "quality_content"
	DeliveredOnTimeReviewPositiveRemarks      ReviewPositiveRemarks = "delivered_on_time"
	ContentPerformedWellReviewPositiveRemarks ReviewPositiveRemarks = "content_performed_well"
	HighEngagementReviewPositiveRemarks       ReviewPositiveRemarks = "high_engagement"
)

func (e ReviewPositiveRemarks) String() string {
	return string(e)
}

func ReviewPositiveRemarksList() []ReviewPositiveRemarks {
	return []ReviewPositiveRemarks{
		RepliesQuicklyReviewPositiveRemarks,
		FollowedInstructionsReviewPositiveRemarks,
		CreativityReviewPositiveRemarks,
		AuthenticContentReviewPositiveRemarks,
		QualityContentReviewPositiveRemarks,
		DeliveredOnTimeReviewPositiveRemarks,
		ContentPerformedWellReviewPositiveRemarks,
		HighEngagementReviewPositiveRemarks,
	}
}

func ReviewPositiveRemarksInfoList() []EnumInfo[ReviewPositiveRemarks] {
	return []EnumInfo[ReviewPositiveRemarks]{
		{Value: RepliesQuicklyReviewPositiveRemarks, Label: "Replies quickly"},
		{Value: FollowedInstructionsReviewPositiveRemarks, Label: "Followed instructions"},
		{Value: CreativityReviewPositiveRemarks, Label: "Creativity"},
		{Value: AuthenticContentReviewPositiveRemarks, Label: "Authentic content"},
		{Value: QualityContentReviewPositiveRemarks, Label: "Quality content"},
		{Value: DeliveredOnTimeReviewPositiveRemarks, Label: "Delivered on time"},
		{Value: ContentPerformedWellReviewPositiveRemarks, Label: "Content performed well"},
		{Value: HighEngagementReviewPositiveRemarks, Label: "High engagement"},
	}
}

type ReviewNegativeRemarks string

const (
	CommunicationReviewNegativeRemarks ReviewNegativeRemarks = "communication"
	SpeedReviewNegativeRemarks         ReviewNegativeRemarks = "speed"
	QualityReviewNegativeRemarks       ReviewNegativeRemarks = "quality"
	CreativityReviewNegativeRemarks    ReviewNegativeRemarks = "creativity"
	AuthenticityReviewNegativeRemarks  ReviewNegativeRemarks = "authenticity"
	ViewCountReviewNegativeRemarks     ReviewNegativeRemarks = "view_count"
	ReachReviewNegativeRemarks         ReviewNegativeRemarks = "reach"
	EngagementReviewNegativeRemarks    ReviewNegativeRemarks = "engagement"
)

func (e ReviewNegativeRemarks) String() string {
	return string(e)
}

func ReviewNegativeRemarksList() []ReviewNegativeRemarks {
	return []ReviewNegativeRemarks{
		CommunicationReviewNegativeRemarks,
		SpeedReviewNegativeRemarks,
		QualityReviewNegativeRemarks,
		CreativityReviewNegativeRemarks,
		AuthenticityReviewNegativeRemarks,
		ViewCountReviewNegativeRemarks,
		ReachReviewNegativeRemarks,
		EngagementReviewNegativeRemarks,
	}
}

func ReviewNegativeRemarksInfoList() []EnumInfo[ReviewNegativeRemarks] {
	return []EnumInfo[ReviewNegativeRemarks]{
		{Value: CommunicationReviewNegativeRemarks, Label: "Communication"},
		{Value: SpeedReviewNegativeRemarks, Label: "Speed"},
		{Value: QualityReviewNegativeRemarks, Label: "Quality"},
		{Value: CreativityReviewNegativeRemarks, Label: "Creativity"},
		{Value: AuthenticityReviewNegativeRemarks, Label: "Authenticity"},
		{Value: ViewCountReviewNegativeRemarks, Label: "View count"},
		{Value: ReachReviewNegativeRemarks, Label: "Reach"},
		{Value: EngagementReviewNegativeRemarks, Label: "Engagement"},
	}
}
