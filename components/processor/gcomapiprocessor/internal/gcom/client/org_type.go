package client

import (
	"fmt"
)

// OrgType enumerates the valid org types that can be used for a grafana.com
// list orgs request
type OrgType string

// Valid GrafanaCloud org types
const (
	OrgTypeAll                    OrgType = "all"
	OrgTypeShared                 OrgType = "shared"
	OrgTypePersonal               OrgType = "personal"
	OrgTypePaid                   OrgType = "paid"
	OrgTypeContracted             OrgType = "contracted"
	OrgTypeSelfServe              OrgType = "self_serve"
	OrgTypeMy                     OrgType = "my"
	OrgTypeGCloud                 OrgType = "gcloud"
	OrgTypeGCP                    OrgType = "gcp"
	OrgTypeAzure                  OrgType = "azure"
	OrgTypeReseller               OrgType = "reseller"
	OrgTypeVIP                    OrgType = "vip"
	OrgTypeFOG                    OrgType = "fog"
	OrgTypeStaff                  OrgType = "staff"
	OrgTypeTrial                  OrgType = "trial"
	OrgTypeGCloudTrial            OrgType = "gcloud-trial"
	OrgTypeGCloudTrialCancelled   OrgType = "gcloud-trial-cancelled"
	OrgTypeAuditPaidUncategorized OrgType = "audit-paid-uncategorized"
)

func (i OrgType) String() string {
	return string(i)
}

func OrgTypeFromString(value string) (OrgType, error) {
	switch value {
	case OrgTypeAll.String():
		return OrgTypeAll, nil
	case OrgTypeShared.String():
		return OrgTypeShared, nil
	case OrgTypePersonal.String():
		return OrgTypePersonal, nil
	case OrgTypePaid.String():
		return OrgTypePaid, nil
	case OrgTypeContracted.String():
		return OrgTypeContracted, nil
	case OrgTypeSelfServe.String():
		return OrgTypeSelfServe, nil
	case OrgTypeMy.String():
		return OrgTypeMy, nil
	case OrgTypeGCloud.String():
		return OrgTypeGCloud, nil
	case OrgTypeGCP.String():
		return OrgTypeGCP, nil
	case OrgTypeAzure.String():
		return OrgTypeAzure, nil
	case OrgTypeReseller.String():
		return OrgTypeReseller, nil
	case OrgTypeVIP.String():
		return OrgTypeVIP, nil
	case OrgTypeFOG.String():
		return OrgTypeFOG, nil
	case OrgTypeStaff.String():
		return OrgTypeStaff, nil
	case OrgTypeTrial.String():
		return OrgTypeTrial, nil
	case OrgTypeGCloudTrial.String():
		return OrgTypeGCloudTrial, nil
	case OrgTypeGCloudTrialCancelled.String():
		return OrgTypeGCloudTrialCancelled, nil
	case OrgTypeAuditPaidUncategorized.String():
		return OrgTypeAuditPaidUncategorized, nil
	}
	return "", fmt.Errorf("org type '%v' is not valid", value)
}
