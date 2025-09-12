package types

const PayloadTypeFallback = "?"

type PayloadType string

const ServerSideAcknowledgeType PayloadType = "ServerSideAcknowledge"
const HealthCheckType PayloadType = "HealthCheck"

func (t PayloadType) String() string {
	return string(t)
}

func PayloadTypeOf[T any]() PayloadType {
	var t T
	return PayloadType(NameOf(t))
}
