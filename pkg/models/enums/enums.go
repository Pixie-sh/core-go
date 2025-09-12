package enums

type ContentTypeStatus string

const (
	ContentTypeStatusActive   ContentTypeStatus = "Active"
	ContentTypeStatusInactive ContentTypeStatus = "Inactive"
	ContentTypeStatusDeleted  ContentTypeStatus = "Deleted"
) //@Field ContentTypeStatus

type Metric string

const (
	ImpressionTargetMetric Metric = "impression"
	ClickedTargetMetric    Metric = "clicked"
	ExitedTargetMetric     Metric = "exited"
	DismissedTargetMetric  Metric = "dismissed"
)

func MetricList() []Metric {
	return []Metric{
		ImpressionTargetMetric,
		ClickedTargetMetric,
		ExitedTargetMetric,
		DismissedTargetMetric,
	}
}

type MetricsTargetType string

const (
	DealsMetricsTargetType    MetricsTargetType = "deals"
	DealApplicationTargetType MetricsTargetType = "deal_applications"
	LocationsTargetType       MetricsTargetType = "locations"
)

func MetricsTargetTypeList() []MetricsTargetType {
	return []MetricsTargetType{
		DealsMetricsTargetType,
		DealApplicationTargetType,
		LocationsTargetType,
	}
}

type ReportedByEntityType string

const (
	ReportedByPartner ReportedByEntityType = "Partner"
)

func (e ReportedByEntityType) String() string {
	return string(e)
}

type ReportedEntityType string

const (
	CreatorReported ReportedEntityType = "Creator"
)

func (e ReportedEntityType) String() string {
	return string(e)
}

type ReportReason string

const (
	NoShowReportReason      ReportReason = "no_show"
	NoReplyReportReason     ReportReason = "no_reply"
	NoContentReportReason   ReportReason = "no_content"
	BadBehaviorReportReason ReportReason = "bad_behavior"
	OtherReportReason       ReportReason = "other"
	LanguageReportReason    ReportReason = "language"
)

func (e ReportReason) String() string {
	return string(e)
}

func ReportReasonList() []ReportReason {
	return []ReportReason{
		NoShowReportReason,
		NoReplyReportReason,
		NoContentReportReason,
		BadBehaviorReportReason,
		OtherReportReason,
		LanguageReportReason,
	}
}

func ReportReasonInfoList() []EnumInfo[ReportReason] {
	return []EnumInfo[ReportReason]{
		{Value: NoShowReportReason, Label: "No Show"},
		{Value: NoReplyReportReason, Label: "No Reply"},
		{Value: NoContentReportReason, Label: "No Content"},
		{Value: BadBehaviorReportReason, Label: "Bad Behavior"},
		{Value: OtherReportReason, Label: "Other"},
		{Value: LanguageReportReason, Label: "Language"},
	}
}
