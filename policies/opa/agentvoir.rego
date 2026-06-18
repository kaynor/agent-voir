package agentvoir.authz

default allow := false

default cache_allowed := false

default semantic_cache_allowed := false

allow if {
  input.agent.lifecycle == "production"
  input.request.provider in input.agent.policies.allowedProviders
}

allow if {
  input.agent.lifecycle == "staging"
  input.environment == "staging"
}

cache_allowed if {
  input.agent.cache.mode != "off"
  not input.request.contains_pii
  not input.request.contains_secret
}

semantic_cache_allowed if {
  cache_allowed
  input.agent.cache.semanticCacheAllowed == true
  count(input.agent.dataClasses) == 0
}

deny[reason] if {
  input.request.contains_pii
  input.agent.policies.piiAllowed == false
  reason := "agent is not approved for PII"
}

deny[reason] if {
  not input.request.provider in input.agent.policies.allowedProviders
  reason := sprintf("provider %s is not approved for this agent", [input.request.provider])
}
