package uidgen

import (
	pulid "github.com/pixie-sh/ulid-go"

	"github.com/pixie-sh/core-go/pkg/uid"
)

var (
	SystemUID        uid.UID
	HealthCheckerUID uid.UID
	BotUID           uid.UID
	BroadcastAllUID  uid.UID

	SystemUUID        string
	HealthCheckerUUID string
	BotUUID           string
	BroadcastAllUUID  string
)

func init() {
	SystemUID, _ = pulid.UnmarshalString("000000048H0000000000000000")
	HealthCheckerUID, _ = pulid.UnmarshalString("000000048H000G000000000000")
	BotUID, _ = pulid.UnmarshalString("00000008H20000000000000000")
	BroadcastAllUID = pulid.EmptyUID

	SystemUUID = SystemUID.UUID()
	HealthCheckerUUID = HealthCheckerUID.UUID()
	BotUUID = BotUID.UUID()
	BroadcastAllUUID = BroadcastAllUID.UUID()
}
