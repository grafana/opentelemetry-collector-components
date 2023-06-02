package client

var hostedPrometheusInstance = `{
  "id": 1,
  "orgId": 2,
  "orgSlug": "example-org",
  "orgName": "example org",
  "type": "prometheus",
  "clusterId": 1,
  "clusterSlug": "us-central1",
  "clusterName": "us-central1",
  "name": "example-org-prom",
  "url": "https://prometheus-us-central1.grafana.net",
  "description": "",
  "plan": "example_plan",
  "billingStartDate": null,
  "billingEndDate": null,
  "billingActiveSeries": 0,
  "billingDpm": 0,
  "billingUsage": 0,
  "trial": 0,
  "trialExpiresAt": null,
  "createdAt": "2018-12-13T02:36:30.000Z",
  "createdBy": "example_user",
  "updatedAt": null,
  "updatedBy": "",
  "links": [
    {
      "rel": "self",
      "href": "/hosted-metrics/1"
    },
    {
      "rel": "org",
      "href": "/orgs/example-org"
    }
  ]
}
`

var hostedGraphiteInstance = `{
  "id": 2,
  "orgId": 2,
  "orgSlug": "example-org",
  "orgName": "example org",
  "type": "graphite",
  "clusterId": 2,
  "clusterSlug": "us-central1",
  "clusterName": "us-central1",
  "name": "example-org-graphite",
  "url": "https://graphite-us-central1.grafana.net",
  "description": "",
  "plan": "example_plan",
  "billingStartDate": null,
  "billingEndDate": null,
  "billingActiveSeries": 0,
  "billingDpm": 0,
  "billingUsage": 0,
  "trial": 0,
  "trialExpiresAt": null,
  "createdAt": "2018-12-13T02:36:30.000Z",
  "createdBy": "example_user",
  "updatedAt": null,
  "updatedBy": "",
  "links": [
    {
      "rel": "self",
      "href": "/hosted-metrics/1"
    },
    {
      "rel": "org",
      "href": "/orgs/example-org"
    }
  ]
}`

var samplePrometheusInstanceParsed = Instance{
	ID:        1,
	OrgID:     2,
	Type:      Prometheus,
	ClusterID: 1,
	Name:      "example-org-prom",
}

var sampleGraphiteInstanceParsed = Instance{
	ID:        2,
	OrgID:     2,
	Type:      Graphite,
	ClusterID: 2,
	Name:      "example-org-graphite",
}

var sampleLogResponse = `{
	"items": [
		{
			"id": 1,
			"orgId": 2,
			"orgSlug": "example_org",
			"orgName": "example org",
			"clusterId": 1,
			"clusterSlug": "us-west1",
			"clusterName": "us-west1",
			"type": "logs",
			"name": "example-org-logs",
			"url": "https://logs-us-west1.grafana.net",
			"plan": "demo",
			"planName": "Demo",
			"description": "",
			"trial": 0,
			"trialExpiresAt": null,
			"createdAt": "2018-12-18T11:48:26.000Z",
			"createdBy": "example_org",
			"updatedAt": null,
			"updatedBy": "",
			"links": [
				{
				"rel": "self",
				"href": "/hosted-logs/1"
				},
				{
				"rel": "org",
				"href": "/orgs/example_org"
				}
			]
		}
	],
	"orderBy": "name",
	"direction": "asc",
	"links": [
		{
		"rel": "self",
		"href": "/hosted-logs"
		}
	]
}`

var sampleLogsParsed = []Instance{
	{
		ID:        1,
		OrgID:     2,
		Type:      Logs,
		ClusterID: 1,
		Name:      "example-org-logs",
	},
}

var sampleAlertsResponse = `{
	"items": [
			{
				"id": 43,
				"orgId": 454533,
				"orgSlug": "alertsgotjosh",
				"orgName": "alertGOTJOSH",
				"orgUrl": "",
				"clusterId": 68,
				"clusterSlug": "alertmanager-us-central1",
				"clusterName": "alertmanager-us-central1",
				"name": "aheadmg-alerts",
				"url": "https://alertmanager-us-central1.grafana.net",
				"description": "",
				"status": "active",
				"createdAt": "2020-07-14T11:59:00.000Z",
				"createdBy": "myself",
				"updatedAt": "2020-07-14T11:59:01.000Z",
				"updatedBy": "mysefl",
				"grafanaInstanceId": 129648,
				"grafanaInstanceStatus": "active",
				"grafanaInstanceName": "gotjosh.grafana.net",
				"grafanaInstanceUrl": "https://gotjosh.grafana.net",
				"grafanaInstanceVersion": "stable",
				"grafanaInstancePlan": "gcloud",
				"hlInstanceId": 5196,
				"hmInstanceGraphiteId": 12584,
				"hmInstancePromId": 12583,
				"links": [
					{
						"rel": "self",
						"href": "/hosted-alerts/43"
					},
					{
						"rel": "org",
						"href": "/orgs/gotjosh1"
					}
				]
			}
		],
		"orderBy": "name",
		"direction": "asc",
		"links": [
			{
				"rel": "self",
				"href": "/hosted-alerts"
			}
		]
}`

var sampleAlertsParsed = []Instance{
	{
		ID:        43,
		OrgID:     454533,
		ClusterID: 68,
		Name:      "aheadmg-alerts",
	},
}

var sampleOrgsResponse = `{
  "items": [
    {
      "id": 1,
      "slug": "raintank",
      "name": "Grafana Labs",
      "url": "https://raintank.io",
      "createdAt": "2015-02-17T14:04:36.000Z",
      "createdBy": "",
      "updatedAt": "2020-07-09T16:39:00.000Z",
      "updatedBy": "",
      "avatar": "custom",
      "links": [
        {
          "rel": "self",
          "href": "/orgs/raintank"
        },
        {
          "rel": "api-keys",
          "href": "/orgs/raintank/api-keys"
        },
        {
          "rel": "members",
          "href": "/orgs/raintank/members"
        }
      ],
      "checksPerMonth": 175320,
      "wpPlan": "",
      "hgInstanceLimit": -1,
      "hmInstanceLimit": 1,
      "hlInstanceLimit": 0,
      "userQuota": -1,
      "supportPlan": "",
      "creditApproved": 0,
      "msaSignedAt": "2017-08-16T11:17:59.000Z",
      "msaSignedBy": "michaelp",
      "enterprisePlugins": 0,
      "grafanaCloud": 21,
      "privacy": "private",
      "reseller": "",
      "emergencySupport": false,
      "gcpAccountId": "",
      "trialStartDate": null,
      "trialEndDate": null,
      "trialLengthDays": null,
      "trialNoticeDate": null,
      "tags": [],
      "accountManagerId": 0,
      "accountManagerUsername": null,
      "accountManagerName": null,
      "accountOwnerId": 0,
      "accountOwnerUsername": null,
      "accountOwnerName": null,
      "hmIncludedSeries": 3000,
      "hmAverageDpm": 6,
      "hmTier1Rate": 16,
      "hmTier2Min": 0,
      "hmTier2Rate": 0,
      "hmTier3Min": 0,
      "hmTier3Rate": 0,
      "hmBillingStartDate": "2018-12-01T00:00:00.000Z",
      "hmBillingEndDate": null,
      "hmBilledToDate": "2020-06-30T23:59:59.000Z",
      "hmOverageWarnDate": null,
      "hmUsage": 7276,
      "hmOverageAmount": 68.42,
      "hmCurrentPrometheusUsage": 2443,
      "hmCurrentGraphiteUsage": 0,
      "hmCurrentUsage": 2443,
      "hlIncludedUsage": 100,
      "hlTier1Rate": 0,
      "hlTier2Min": 0,
      "hlTier2Rate": 0,
      "hlBillingStartDate": null,
      "hlBillingEndDate": null,
      "hlBilledToDate": null,
      "hlOverageWarnDate": null,
      "hlUsage": 4.87,
      "hlOverageAmount": 10.34,
      "hlCurrentUsage": 0,
      "hgIncludedUsers": 0,
      "hgTier1Rate": 0,
      "hgTier2Min": 0,
      "hgTier2Rate": 0,
      "hgBillingStartDate": null,
      "hgBillingEndDate": null,
      "hgBilledToDate": null,
      "hgOverageWarnDate": null,
      "hgActiveUsers": 0,
      "hgOverageAmount": 0,
      "hgCurrentActiveUsers": 1,
      "hgDatasourceCnts": {
        "alexanderzobnin-zabbix-datasource": 1,
        "elasticsearch": 1,
        "graphite": 1,
        "mysql": 1
      },
      "totalOverageAmount": 68.42,
      "memberCnt": 4,
      "licenseCnt": 0,
      "licenseConfiguredCnt": 0,
      "licenseUnconfiguredCnt": 0,
      "hmInstanceCnt": 1,
      "hmGraphiteInstanceCnt": 0,
      "hmPrometheusInstanceCnt": 1,
      "hgInstanceCnt": 1,
      "hlInstanceCnt": 0,
      "ubersmithClientId": 30825,
      "committedArr": 468,
      "zendeskId": 360289420451,
      "salesforceAccountId": "",
      "salesforceLeadId": "",
      "happinessRating": null,
      "happinessNote": null,
      "happinessCreatedAt": null,
      "happinessExpiredAt": null,
      "happinessChangedAt": null,
      "happinessUserName": null,
      "cancellationClientNotes": null,
      "cancellationNotes": null,
      "cancellationReason": "",
      "netPromoterScore": null,
      "estimatedArr": 1289.04,
      "contractType": "self_serve"
    }
  ],
  "total": 1,
  "pages": 1,
  "pageSize": 1000000,
  "page": 1,
  "orderBy": "name",
  "direction": "asc",
  "links": [
    {
      "rel": "self",
      "href": "/orgs"
    }
  ]
}`

var sampleOrgsParsed = []Org{
	{
		ID:                    1,
		Slug:                  "raintank",
		Name:                  "Grafana Labs",
		GrafanaCloudType:      OrgGrafanaCloudTypeGCPSubscribed,
		ContractType:          OrgContractTypeSelfServe,
		MetricsUsage:          7276,
		MetricsOverageAmount:  68.42,
		MetricsIncludedSeries: 3000,
		LogsUsage:             4.87,
		LogsOverageAmount:     10.34,
		LogsIncludedUsage:     100,
	},
}

var sampleGrafanaResponse = `{
  "items":[
    {
      "id":12,
      "orgId":1,
      "orgSlug":"testorg",
      "orgName":"Test Org",
      "name":"teststack",
      "url":"https://org_one.grafana.net",
      "slug":"teststack",
      "version":"stable",
      "description":"",
      "status":"active",
      "createdAt":"2017-08-10T23:25:10.000Z",
      "createdBy":"dcech",
      "updatedAt":"2021-02-25T20:21:55.000Z",
      "updatedBy":"dcech",
      "trial":0,
      "trialExpiresAt":null,
      "clusterId":57,
      "clusterSlug":"hg-free-us-central1",
      "clusterName":"HG Free US Central1",
      "plan":"gcloud",
      "planName":"Grafana Cloud",
      "billingStartDate":"2021-02-04T21:13:22.000Z",
      "billingEndDate":null,
      "billingActiveUsers":0,
      "currentActiveUsers":0,
      "currentActiveAdminUsers":0,
      "currentActiveEditorUsers":0,
      "currentActiveViewerUsers":0,
      "datasourceCnts":{
        "prometheus":1
      },
      "userQuota":10,
      "dashboardQuota":-1,
      "alertQuota":-1,
      "ssl":true,
      "customAuth":true,
      "customDomain":true,
      "support":true,
      "runningVersion":"7.4.3 (commit: 010f20c1c8, branch: HEAD)",
      "hmInstancePromId":5727,
      "hmInstancePromUrl":"https://prometheus-us-central1.grafana.net",
      "hmInstancePromName":"teststack-prom",
      "hmInstancePromStatus":"active",
      "hmInstancePromCurrentUsage":1577,
      "hmInstanceGraphiteId":5178,
      "hmInstanceGraphiteUrl":"https://graphite-us-central1.grafana.net",
      "hmInstanceGraphiteName":"teststack-graphite",
      "hmInstanceGraphiteStatus":"active",
      "hmInstanceGraphiteCurrentUsage":2,
      "hlInstanceId":34,
      "hlInstanceUrl":"https://logs-prod-us-central1.grafana.net",
      "hlInstanceName":"teststack-logs",
      "hlInstanceStatus":"active",
      "hlInstanceCurrentUsage":0,
      "amInstanceId":1444,
      "amInstanceName":"teststack-alerts",
      "amInstanceStatus":"active",
      "amInstanceGeneratorUrl":"https://teststack.grafana.net",
      "htInstanceId":5007,
      "htInstanceName":"teststack-traces",
      "htInstanceStatus":"active",
      "links":[
        {
          "rel":"self",
          "href":"/instances/teststack"
        },
        {
          "rel":"org",
          "href":"/orgs/teststack"
        },
        {
          "rel":"plugins",
          "href":"/instances/teststack/plugins"
        }
      ],
      "orgAccountManagerName":null,
      "orgAccountOwnerName":null,
      "hCreatedAt":null,
      "hNote":null
    }
  ]
}`

var sampleGrafanaParsed = []Instance{
	{
		ID:        12,
		OrgID:     1,
		ClusterID: 57,
		Name:      "teststack",
	},
}

var sampleHMClusterResponse = `{
  "items": [
    {
      "id": 14,
      "slug": "alertmanager-dev-us-central-0",
      "name": "dev-us-central-0.alertmanager",
      "url": "https://alertmanager-dev-us-central1.grafana-dev.net",
      "deployTo": false,
      "description": "dev-us-central-0.alertmanager",
      "regionId": 0,
      "regionSlug": null,
      "regionName": null,
      "createdAt": "2022-07-26T22:37:00.000Z",
      "createdBy": "raintank-81",
      "updatedAt": null,
      "updatedBy": "",
      "links": [
        {
          "rel": "self",
          "href": "/hm-clusters/alertmanager-dev-us-central-0"
        },
        {
          "rel": "instances",
          "href": "/instances?cluster=alertmanager-dev-us-central-0"
        }
      ]
    }
  ],
  "orderBy": "name",
  "direction": "asc",
  "total": 1,
  "pages": 1,
  "pageSize": 1000000,
  "page": 1,
  "links": [
    {
      "rel": "self",
      "href": "/hm-clusters"
    }
  ]
}`

var sampleHMClusterParsed = Cluster{
	ID:   14,
	Slug: "alertmanager-dev-us-central-0",
}

var sampleHGClusterResponse = `{
  "items": [
    {
      "id": 1,
      "slug": "dev-us-central-0",
      "name": "dev-us",
      "url": "https://hg-api-dev-us-central-0.grafana.net",
      "deployTo": "both",
      "description": "",
      "regionId": 1,
      "regionSlug": "dev-us-central",
      "regionName": "dev-us-central",
      "createdAt": "2022-06-20T16:08:51.000Z",
      "createdBy": "grafana",
      "updatedAt": "2023-01-13T18:43:53.000Z",
      "updatedBy": "",
      "links": [
        {
          "rel": "self",
          "href": "/hg-clusters/dev-us-central-0"
        },
        {
          "rel": "instances",
          "href": "/instances?cluster=dev-us-central-0"
        }
      ]
    }
  ],
  "orderBy": "name",
  "direction": "asc",
  "total": 1,
  "pages": 1,
  "pageSize": 1000000,
  "page": 1,
  "links": [
    {
      "rel": "self",
      "href": "/hg-clusters"
    }
  ]
}`

var sampleHGClusterParsed = Cluster{
	ID:   1,
	Slug: "dev-us-central-0",
}
