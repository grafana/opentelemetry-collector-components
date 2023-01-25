package client

// OrgGrafanaCloudType enumerates the valid grafana cloud subscription types for grafana.com
type OrgGrafanaCloudType int

// Valid GrafanaCloud subscription types
const (
	OrgGrafanaCloudTypeDisabled                  OrgGrafanaCloudType = 0
	OrgGrafanaCloudTypeSubscribed                OrgGrafanaCloudType = 1
	OrgGrafanaCloudTypePendingCancellation       OrgGrafanaCloudType = 2
	OrgGrafanaCloudTypeCancelled                 OrgGrafanaCloudType = 3
	OrgGrafanaCloudTypeTrial                     OrgGrafanaCloudType = 4
	OrgGrafanaCloudTypeTrialCancelled            OrgGrafanaCloudType = 5
	OrgGrafanaCloudTypeTrialAddedCC              OrgGrafanaCloudType = 6
	OrgGrafanaCloudTypeFree                      OrgGrafanaCloudType = 9
	OrgGrafanaCloudTypeLegacySubscribed          OrgGrafanaCloudType = 11
	OrgGrafanaCloudTypeLegacyPendingCancellation OrgGrafanaCloudType = 12
	OrgGrafanaCloudTypeLegacyCancelled           OrgGrafanaCloudType = 13
	OrgGrafanaCloudTypeGCPFlatFeeSubscribed      OrgGrafanaCloudType = 14
	OrgGrafanaCloudTypeGCPFlatFeeOnHold          OrgGrafanaCloudType = 15
	OrgGrafanaCloudTypeGCPSubscribed             OrgGrafanaCloudType = 21
	OrgGrafanaCloudTypeGCPPending                OrgGrafanaCloudType = 22
	OrgGrafanaCloudTypeGCPCancelled              OrgGrafanaCloudType = 23
	OrgGrafanaCloudTypeAzureSubscribed           OrgGrafanaCloudType = 31
	OrgGrafanaCloudTypeAzureSuspended            OrgGrafanaCloudType = 32
	OrgGrafanaCloudTypeAzureCancelled            OrgGrafanaCloudType = 33
)

// OrgGrafanaCloudSubscribedTypes contains GrafanaCloud subscription types that
// are actively subscribed to Grafana Cloud
var OrgGrafanaCloudSubscribedTypes = []OrgGrafanaCloudType{
	OrgGrafanaCloudTypeSubscribed,
	OrgGrafanaCloudTypePendingCancellation,
	OrgGrafanaCloudTypeTrial,
	OrgGrafanaCloudTypeTrialAddedCC,
	OrgGrafanaCloudTypeLegacySubscribed,
	OrgGrafanaCloudTypeLegacyPendingCancellation,
	OrgGrafanaCloudTypeGCPFlatFeeSubscribed,
	OrgGrafanaCloudTypeGCPSubscribed,
	OrgGrafanaCloudTypeAzureSubscribed,
}
