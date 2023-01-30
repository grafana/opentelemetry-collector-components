package client

import (
	"fmt"
)

// OrgContractType enumerates the valid org contract types for grafana.com
type OrgContractType string

// Valid GrafanaCloud instance types
const (
	OrgContractTypeNone       OrgContractType = "none"
	OrgContractTypeSelfServe  OrgContractType = "self_serve"
	OrgContractTypeContracted OrgContractType = "contracted"
)

func (i OrgContractType) String() string {
	return string(i)
}

func OrgContractTypeFromString(value string) (OrgContractType, error) {
	switch value {
	case OrgContractTypeNone.String():
		return OrgContractTypeNone, nil
	case OrgContractTypeSelfServe.String():
		return OrgContractTypeSelfServe, nil
	case OrgContractTypeContracted.String():
		return OrgContractTypeContracted, nil
	}
	return "", fmt.Errorf("org contract type '%v' is not valid", value)
}
