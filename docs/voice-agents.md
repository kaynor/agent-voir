considering agents are good for triage and troubleshooting, imagine a scenario where an agent is assigned for responding to pagerduty calls. This  agent can respond back and hop onto phone calls talking to real human beings and at same time troubleshooting the issue. Agent might potentially ring up another human if agent is not able to resolve. Is there any metadata that needs to be captured here or for any other similar scenario?



Yes — this scenario needs **extra metadata** beyond normal agent registry fields.

A PagerDuty/on-call troubleshooting agent is not just a “chat agent.” It is a **real-time operational responder** that can talk to humans, join calls, inspect systems, trigger workflows, escalate, and possibly affect production. So AgentVoir should model this as a special class:

> **Incident Response Agent**
> or
> **Operational Responder Agent**

PagerDuty-style systems already revolve around incident assignment, escalation policies, schedules, event routing/orchestration, and incident lifecycle stages such as detect, triage, diagnose, remediate, and learn. PagerDuty docs describe escalation policies as rules that move incidents to the next responder if the current responder does not acknowledge in time, and Event Orchestration as routing events through rules based on event content. ([PagerDuty][1])

---

# 1. New metadata category: incident-response authority

You need to capture **what the agent is allowed to do during an incident**.

```yaml
incident_response_profile:
  enabled: true
  responder_role: "first_line_triage"
  incident_platforms:
    - "pagerduty"
    - "slack"
    - "zoom"
    - "servicenow"

  allowed_incident_actions:
    - "acknowledge_incident"
    - "add_incident_note"
    - "join_war_room"
    - "summarize_logs"
    - "page_human_responder"
    - "create_status_update"

  forbidden_incident_actions:
    - "resolve_incident_without_human"
    - "restart_production_service"
    - "rollback_release"
    - "disable_security_controls"

  max_autonomy_level: "triage_only"
```

Possible autonomy levels:

```text
observe_only
triage_only
recommend_actions
execute_low_risk_actions
execute_with_approval
full_autonomous_response
```

For production incidents, I would start AgentVoir with **observe_only**, **triage_only**, and **recommend_actions** before allowing execution.

---

# 2. Escalation and paging metadata

The agent must know **when to call a human**, **who to call**, and **how long to wait**.

PagerDuty incidents are assigned through escalation policies, and if the assigned user does not acknowledge before the escalation timeout, the incident escalates to the next level. ([PagerDuty][2])

```yaml
escalation_policy:
  source: "pagerduty"
  policy_id: "pd-policy-payments-prod"
  service_id: "payments-api-prod"

  agent_escalation_rules:
    - condition: "severity == 'SEV1'"
      action: "page_primary_oncall_immediately"

    - condition: "agent_confidence < 0.70"
      action: "page_human_responder"

    - condition: "incident_age_minutes > 10 and not mitigated"
      action: "escalate_to_secondary"

    - condition: "customer_impact_detected == true"
      action: "notify_incident_commander"

  max_agent_triage_minutes: 7
  require_human_for_resolution: true
```

This lets AgentVoir answer:

```text
When is the agent allowed to keep troubleshooting?
When must it page a human?
Who did it page?
Did it follow the escalation policy?
Did it delay escalation too long?
```

---

# 3. Human communication metadata

Once the agent talks to real humans, AgentVoir needs metadata for **communication identity, channel, disclosure, call recording, transcript storage, and audit**.

```yaml
human_communication:
  channels:
    - type: "voice_call"
      provider: "twilio"
      phone_number_id: "voice-line-prod-01"
      caller_id_name: "AgentVoir Incident Assistant"

    - type: "slack"
      workspace: "company-prod"
      allowed_channels:
        - "#incident-war-room"
        - "#sre-oncall"

    - type: "zoom"
      join_allowed: true
      speak_allowed: true

  voice_identity:
    tts_provider: "elevenlabs_or_internal"
    voice_id: "approved-incident-agent-voice"
    ai_disclosure_required: true
    opening_script_id: "incident-agent-disclosure-v1"

  recording:
    recording_allowed: true
    consent_required: true
    transcript_enabled: true
    transcript_retention_days: 30
    pii_redaction_required: true
```

For voice calls, this is especially important. Twilio’s guidance says recording-consent laws vary and recommends getting consent from all participants before recording, with a clear explanation that the call may or will be recorded. ([Twilio Help Center][3])

AgentVoir should capture:

```text
Was this call recorded?
Was consent obtained?
Who was on the call?
Was the AI identity disclosed?
Where is the transcript stored?
Was PII redacted?
Can this transcript be used for eval/training?
```

---

# 4. Real-time voice pipeline metadata

A voice incident agent has a full speech pipeline, not just an LLM.

You should capture dependencies for:

```text
ASR / speech-to-text model
LLM reasoning model
TTS / text-to-speech model
telephony provider
conference provider
transcription provider
noise suppression
speaker diarization
```

Example:

```yaml
voice_pipeline:
  speech_to_text:
    provider: "deepgram"
    model: "nova-3"
    p95_latency_ms: 450
    confidence_threshold: 0.82

  reasoning_model:
    provider: "openai"
    model: "gpt-4.1"
    p95_latency_ms: 2800

  text_to_speech:
    provider: "internal-tts"
    voice_id: "incident-agent-voice-v2"
    p95_latency_ms: 600

  conversation_metrics:
    avg_turn_latency_ms: 1800
    interruption_detection_enabled: true
    barge_in_supported: true
    silence_timeout_seconds: 6
```

This matters because a voice agent can fail because of bad ASR, slow TTS, call drops, or misunderstood speech — not just bad reasoning.

---

# 5. Runbook and troubleshooting metadata

For this type of agent, you should capture which runbooks it knows and which ones it is allowed to execute.

```yaml
runbook_access:
  allowed_runbooks:
    - id: "runbook:payments-api-5xx"
      version: "v12"
      execution_mode: "recommend_only"

    - id: "runbook:kafka-consumer-lag"
      version: "v7"
      execution_mode: "execute_low_risk_steps"

    - id: "runbook:database-failover"
      version: "v5"
      execution_mode: "human_approval_required"

  forbidden_runbooks:
    - "runbook:manual-prod-db-write"
    - "runbook:disable-fraud-rules"

  runbook_adherence_required: true
```

Execution modes:

```text
read_only
recommend_only
execute_safe_checks
execute_low_risk_steps
approval_required
forbidden
```

---

# 6. Production action metadata

The biggest risk is not that the agent talks. The risk is that it **does something**.

So each action needs a safety classification.

```yaml
production_action_policy:
  allowed_tools:
    - name: "read_datadog_metrics"
      risk_level: "low"
      approval_required: false

    - name: "query_logs"
      risk_level: "low"
      approval_required: false

    - name: "restart_service"
      risk_level: "high"
      approval_required: true

    - name: "rollback_deployment"
      risk_level: "critical"
      approval_required: true
      approver_role: "incident_commander"

    - name: "disable_feature_flag"
      risk_level: "medium"
      approval_required: true

  global_restrictions:
    require_human_approval_for_customer_impacting_actions: true
    require_human_approval_for_data_modification: true
    require_human_approval_for_security_changes: true
```

AgentVoir should store every attempted action:

```yaml
incident_action_log:
  incident_id: "INC-12345"
  agent_id: "agent:prod-incident-responder"
  action: "restart_service"
  target: "payments-api"
  proposed_at: "2026-06-19T03:14:00Z"
  approved_by: "sre-oncall-user"
  executed_at: "2026-06-19T03:15:00Z"
  result: "success"
  rollback_available: true
```

This becomes vital for audit and postmortem.

---

# 7. Confidence and uncertainty metadata

For incidents, AgentVoir should force the agent to expose uncertainty.

```yaml
diagnostic_confidence:
  current_hypothesis: "database connection pool exhaustion"
  confidence_score: 0.74
  evidence:
    - "p95 latency increased after deploy"
    - "db connection pool saturation at 98%"
    - "error rate concentrated in payments-api"

  uncertainty_reasons:
    - "logs are incomplete for one region"
    - "recent deploy metadata unavailable"

  escalation_required_if_confidence_below: 0.70
```

This is important because an agent should not keep confidently talking during an outage if it is guessing.

---

# 8. Human handoff metadata

When the agent escalates, it should hand off context cleanly.

```yaml
handoff:
  handoff_required: true
  handoff_target_role: "primary_oncall"
  handoff_format: "incident_brief"

  handoff_brief_fields:
    - "incident_summary"
    - "customer_impact"
    - "timeline"
    - "suspected_cause"
    - "evidence_collected"
    - "actions_taken"
    - "actions_recommended"
    - "open_questions"
    - "risk_level"
```

Example handoff record:

```yaml
human_handoff_record:
  incident_id: "INC-12345"
  from_agent: "agent:prod-incident-responder"
  to_human: "primary-oncall"
  reason: "agent_confidence_below_threshold"
  handoff_time: "2026-06-19T03:21:00Z"
  summary: "Payments API 5xx spike. Suspected DB pool exhaustion. No production changes executed."
```

This is a major feature. A good incident agent should not just page a human; it should page a human with a useful summary.

---

# 9. War-room behavior metadata

If the agent joins Slack, Zoom, or phone bridges, you need rules for how it behaves.

```yaml
war_room_behavior:
  can_join_meeting: true
  can_speak_unprompted: false
  can_answer_direct_questions: true
  can_interrupt: false
  can_summarize_periodically: true
  summary_interval_minutes: 10

  allowed_speech_modes:
    - "answer_questions"
    - "provide_status_summary"
    - "recommend_next_step"

  forbidden_speech_modes:
    - "assign_blame"
    - "declare_incident_resolved_without_human"
    - "communicate_to_customers_without_approval"
```

This is important because “being in a call” has social and operational consequences.

---

# 10. Incident communication policy

For incidents, the agent may produce messages for internal teams, executives, customers, or regulators. These need different approval rules.

```yaml
incident_comms_policy:
  internal_status_updates:
    allowed: true
    approval_required: false

  executive_updates:
    allowed: true
    approval_required: true
    approver_role: "incident_commander"

  customer_updates:
    allowed: "draft_only"
    approval_required: true
    approver_role: "communications_lead"

  regulatory_updates:
    allowed: "draft_only"
    approval_required: true
    approver_role: "legal_compliance"
```

This is especially important in finance, healthcare, infrastructure, or enterprise SaaS.

---

# 11. Incident quality and performance metadata

For this agent type, generic quality score is not enough. You need incident-specific metrics.

```yaml
incident_agent_metrics:
  mtta_seconds: 42
  mtte_seconds: 180
  mttr_contribution_minutes: 12

  triage_accuracy_score: 8.6
  escalation_accuracy_score: 9.1
  false_escalation_rate: 0.04
  missed_escalation_rate: 0.01

  human_override_rate: 0.18
  runbook_adherence_score: 0.93
  unsafe_action_attempts_30d: 0
  customer_comms_approval_rejections_30d: 2
```

Important metrics:

```text
Mean time to acknowledge
Mean time to useful summary
Mean time to escalation
Mean time to mitigation recommendation
Correct escalation rate
False escalation rate
Missed escalation rate
Human override rate
Runbook adherence
Unsafe action attempts
Postmortem usefulness score
```

---

# 12. Post-incident learning metadata

AgentVoir should connect incident traces back into evals and runbook improvements.

```yaml
post_incident_learning:
  generate_postmortem_draft: true
  convert_incident_to_eval_candidate: true
  require_human_review_for_eval: true

  captured_artifacts:
    - "timeline"
    - "call_transcript"
    - "slack_messages"
    - "logs_queries"
    - "metric_snapshots"
    - "agent_decision_log"
    - "human_overrides"
    - "final_resolution"

  learning_outputs:
    - "new_eval_case"
    - "runbook_update_suggestion"
    - "policy_update_suggestion"
    - "tool_permission_review"
```

This is where AgentVoir becomes extremely valuable: every incident can improve future agent behavior.

---

# 13. New entities AgentVoir should add

For this scenario, I would add these entities:

```text
IncidentResponderProfile
IncidentSession
IncidentActionLog
CommunicationSession
VoiceCallTranscript
EscalationDecision
HumanHandoffRecord
RunbookBinding
WarRoomParticipation
IncidentCommsPolicy
IncidentPostmortemArtifact
IncidentQualityScore
```

You can model this as an extension of the normal agent schema.

---

# 14. Example full metadata snippet

```yaml
agent:
  id: "agent:prod-incident-responder"
  name: "Production Incident Responder Agent"
  type: "incident_response_agent"
  owner_team: "SRE Platform"
  risk_tier: "critical"
  lifecycle_stage: "prod"

incident_response_profile:
  responder_role: "first_line_triage"
  max_autonomy_level: "recommend_actions"
  require_human_for_resolution: true
  max_agent_triage_minutes: 7

pagerduty:
  service_ids:
    - "payments-api-prod"
    - "checkout-prod"
  escalation_policy_id: "pd-policy-sre-primary"
  can_acknowledge: true
  can_resolve: false
  can_reassign: true
  can_escalate: true

human_communication:
  voice_enabled: true
  ai_disclosure_required: true
  recording_allowed: true
  consent_required: true
  transcript_enabled: true
  pii_redaction_required: true
  allowed_channels:
    - "phone"
    - "slack"
    - "zoom"

war_room_behavior:
  can_join_meeting: true
  can_speak_unprompted: false
  can_answer_direct_questions: true
  can_summarize_periodically: true
  summary_interval_minutes: 10

runbook_access:
  allowed_runbooks:
    - id: "runbook:payments-api-5xx"
      version: "v12"
      execution_mode: "recommend_only"
    - id: "runbook:kafka-consumer-lag"
      version: "v7"
      execution_mode: "execute_safe_checks"

production_action_policy:
  default_mode: "recommend_only"
  approval_required_for:
    - "restart_service"
    - "rollback_deployment"
    - "disable_feature_flag"
    - "modify_database"
    - "send_customer_update"

escalation_policy:
  escalate_if:
    - "severity == SEV1"
    - "agent_confidence < 0.70"
    - "customer_impact_detected == true"
    - "incident_age_minutes > 10 and not mitigated"

observability:
  trace_every_incident: true
  record_decision_log: true
  capture_tool_calls: true
  capture_model_calls: true
  dashboard_url: "grafana://agentvoir/incident-responder"

quality:
  incident_specific_metrics:
    triage_accuracy_score: 8.7
    escalation_accuracy_score: 9.2
    runbook_adherence_score: 0.94
    human_override_rate: 0.16
```

---

## Similar scenarios that need the same metadata pattern

This is not only for PagerDuty. The same metadata model applies to:

| Scenario                              | Special concern                                   |
| ------------------------------------- | ------------------------------------------------- |
| **SOC/security incident agent**       | May isolate hosts, block IPs, disable accounts    |
| **Trading operations agent**          | May contact traders, check limits, halt workflows |
| **Customer support escalation agent** | May call customers or issue credits               |
| **Healthcare triage agent**           | May speak with patients or escalate to clinicians |
| **DevOps release agent**              | May rollback, restart, or freeze deployments      |
| **Finance operations agent**          | May approve payments or flag fraud                |
| **Legal/compliance agent**            | May draft regulated communications                |

The common pattern is:

> Any agent that can communicate with humans in real time or trigger operational changes needs **authority, communication, consent, escalation, audit, and handoff metadata**.

---

## My recommendation

For AgentVoir, create a special metadata extension called:

```text
OperationalAgentProfile
```

Under that, support subtypes:

```text
incident_responder
security_responder
customer_support_responder
trading_ops_responder
healthcare_triage_responder
release_manager
```

The most important fields are:

```text
1. What incidents/events can activate this agent?
2. What systems can it inspect?
3. What humans can it contact?
4. What channels can it speak on?
5. Is AI disclosure required?
6. Can calls be recorded/transcribed?
7. What actions can it execute?
8. What actions require approval?
9. When must it escalate?
10. How is handoff recorded?
11. How are decisions audited?
12. How does the incident become an eval/test afterward?
```

This would be a very strong addition to AgentVoir because it moves the product beyond “agent registry” into **enterprise operational governance for autonomous agents**.

[1]: https://support.pagerduty.com/main/docs/escalation-policies?utm_source=chatgpt.com "Escalation Policy Basics"
[2]: https://support.pagerduty.com/main/docs/incidents?utm_source=chatgpt.com "Incidents"
[3]: https://help.twilio.com/articles/360011522553-Legal-Considerations-with-Recording-Voice-and-Video-Communications?utm_source=chatgpt.com "Legal Considerations with Recording Voice and Video ..."
