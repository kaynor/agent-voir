package agentvoir.authz_test

import rego.v1

import data.agentvoir.authz

test_production_agent_allowed_provider if {
	authz.allow with input as {
		"agent": {
			"lifecycle": "production",
			"policies": {"allowedProviders": ["openai", "anthropic"]},
			"cache": {"mode": "exact_only", "semanticCacheAllowed": false},
			"dataClasses": [],
		},
		"request": {"provider": "openai", "contains_pii": false, "contains_secret": false},
		"environment": "production",
	}
}

test_production_agent_denied_provider if {
	not authz.allow with input as {
		"agent": {
			"lifecycle": "production",
			"policies": {"allowedProviders": ["openai"]},
			"cache": {"mode": "exact_only", "semanticCacheAllowed": false},
			"dataClasses": [],
		},
		"request": {"provider": "anthropic", "contains_pii": false, "contains_secret": false},
		"environment": "production",
	}
}

test_staging_agent_allowed_in_staging if {
	authz.allow with input as {
		"agent": {"lifecycle": "staging", "cache": {"mode": "exact_only", "semanticCacheAllowed": false}, "dataClasses": []},
		"request": {"provider": "openai", "contains_pii": false, "contains_secret": false},
		"environment": "staging",
	}
}

test_cache_allowed_without_pii if {
	authz.cache_allowed with input as {
		"agent": {"cache": {"mode": "exact_only", "semanticCacheAllowed": true}, "dataClasses": []},
		"request": {"contains_pii": false, "contains_secret": false},
	}
}

test_deny_pii_when_not_allowed if {
	"agent is not approved for PII" in authz.deny with input as {
		"agent": {"policies": {"piiAllowed": false, "allowedProviders": ["openai"]}},
		"request": {"provider": "openai", "contains_pii": true},
	}
}
