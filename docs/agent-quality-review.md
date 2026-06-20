Are we already capturing feedback about quality of agent? Maybe everyday humans and other agents should be tasked with rating agents from scale of 1 to 10 on output. 


Based on what we discussed so far, **we are capturing eval metadata and observability**, but we have **not yet explicitly defined a human/agent feedback system for quality scoring**.

So yes, this should be added. It is a very good feature for AgentVoir.

I would call it something like:

> **Agent Quality Feedback Loop**
> or
> **Agent Reputation & Quality Scoring**

The idea: every agent should accumulate quality signals from **humans, automated evaluators, and other agents**.

---

## Why this matters

Token usage and cost tell you:

> “How expensive is this agent?”

Traces tell you:

> “What did this agent do?”

Evals tell you:

> “Did this agent pass known test cases?”

But feedback tells you:

> “Was this agent actually useful in the real world?”

That is a major missing piece.

---

# 1. Add quality metadata to each agent

Each agent should have a quality profile.

```yaml
quality:
  current_score: 8.4
  score_scale: "1-10"
  confidence: 0.78
  rating_count_30d: 245
  human_rating_count_30d: 91
  agent_rating_count_30d: 154
  last_rated_at: "2026-06-18T15:30:00Z"

  trend:
    7d_avg: 8.6
    30d_avg: 8.4
    90d_avg: 8.1

  status: "healthy"
```

Useful statuses:

```text
healthy
watch
degraded
quarantined
disabled
needs_review
```

---

# 2. Do not use only one 1–10 score

A simple 1–10 rating is useful, but it is too vague by itself.

Instead, capture both:

```text
Overall score: 1–10
```

and sub-scores:

```text
Accuracy
Usefulness
Completeness
Grounding
Safety
Policy compliance
Latency satisfaction
Format correctness
Tool-use correctness
```

Example:

```yaml
rating:
  overall_score: 8
  dimensions:
    accuracy: 9
    usefulness: 8
    completeness: 7
    grounding: 9
    safety: 10
    format_correctness: 8
    tool_use_correctness: 9
```

This helps you understand *why* an agent is rated poorly.

For example:

```text
Agent A: overall 6/10 because it is slow.
Agent B: overall 6/10 because it hallucinates.
Agent C: overall 6/10 because it uses the wrong tool.
```

Those are very different problems.

---

# 3. Capture human feedback

For humans, make feedback easy and lightweight.

After each agent output, allow:

```text
👍 / 👎
1–10 score
Optional reason
Issue category
```

Possible categories:

```text
wrong_answer
incomplete_answer
hallucination
bad_format
policy_violation
slow_response
wrong_tool_used
unhelpful
unsafe
excellent_output
```

Example schema:

```yaml
human_feedback:
  feedback_id: "fb_123"
  agent_id: "agent:position-limit-checker"
  run_id: "run_789"
  user_id: "user_456"
  score: 8
  thumbs: "up"
  category: "incomplete_answer"
  comment: "Good analysis but missing CME source citation."
  created_at: "2026-06-18T16:00:00Z"
```

Important: this should link back to the **specific run/trace**, not just the agent.

---

# 4. Capture agent-generated feedback

Other agents can review outputs too.

For example:


| Reviewer agent                  | What it checks                                |
| ------------------------------- | --------------------------------------------- |
| **Grounding Reviewer Agent**    | Are claims supported by sources?              |
| **Policy Reviewer Agent**       | Did output violate internal policy?           |
| **Format Reviewer Agent**       | Did output match required JSON/report format? |
| **Safety Reviewer Agent**       | Was there risky or prohibited behavior?       |
| **Cost Reviewer Agent**         | Was the expensive model necessary?            |
| **Tool-use Reviewer Agent**     | Did the agent call the right tools?           |
| **Completeness Reviewer Agent** | Did the answer satisfy the task?              |


Example:

```yaml
agent_feedback:
  reviewer_agent_id: "agent:grounding-reviewer"
  target_agent_id: "agent:fundamental-analysis-agent"
  run_id: "run_abc"
  score: 7
  dimensions:
    factual_grounding: 6
    citation_quality: 7
    completeness: 8
  findings:
    - "Two claims were not backed by retrieved evidence."
    - "Output was mostly useful but needs stronger source attribution."
```

This is powerful because you can evaluate many runs without requiring humans every time.

---

# 5. Daily rating jobs make sense

Your idea of daily ratings is strong.

AgentVoir could support scheduled quality jobs like:

```yaml
quality_review_schedule:
  enabled: true
  frequency: "daily"
  sample_strategy: "risk_weighted"
  sample_size_per_day: 50
  reviewers:
    human_reviewers:
      enabled: true
      sample_size: 10
    agent_reviewers:
      enabled: true
      sample_size: 50
```

But I would not ask humans to rate every single output. Instead:

1. **Humans rate high-risk samples**
2. **Agents rate a larger daily sample**
3. **Users can rate outputs naturally during usage**
4. **Bad outputs automatically enter eval datasets**

---

# 6. Use risk-weighted sampling

Not every agent needs the same review frequency.

Example:


| Agent risk tier | Review strategy                         |
| --------------- | --------------------------------------- |
| Low risk        | Weekly sample                           |
| Medium risk     | Daily automated review                  |
| High risk       | Daily automated + human sample          |
| Critical risk   | Human approval or frequent human review |


For AgentVoir:

```yaml
quality_sampling:
  strategy: "risk_weighted"
  rules:
    - risk_tier: "low"
      human_reviews_per_week: 5
      agent_reviews_per_day: 10

    - risk_tier: "medium"
      human_reviews_per_week: 10
      agent_reviews_per_day: 50

    - risk_tier: "high"
      human_reviews_per_day: 10
      agent_reviews_per_day: 100

    - risk_tier: "critical"
      human_reviews_per_day: 25
      agent_reviews_per_day: 250
```

This is better than reviewing everything equally.

---

# 7. Add quality gates

Feedback should not just be displayed. It should affect runtime behavior.

Example rules:

```text
If quality score drops below 7.0, mark agent as "watch."
If score drops below 6.0, require human approval for risky tool calls.
If score drops below 5.0, quarantine the agent.
If safety score drops below threshold, disable external actions.
If hallucination rate rises, force cheaper model off and route to stronger model.
```

Example:

```yaml
quality_gates:
  - condition: "quality.7d_avg < 7.0"
    action: "mark_watch"

  - condition: "quality.7d_avg < 6.0"
    action: "require_human_approval"

  - condition: "safety_score < 8.0"
    action: "disable_external_tools"

  - condition: "quality.7d_avg < 5.0"
    action: "quarantine_agent"
```

This turns quality feedback into governance.

---

# 8. Connect feedback to evals

This is very important.

When a human gives bad feedback, AgentVoir should ask:

> Should this become a regression test?

Example:

```yaml
feedback_to_eval:
  enabled: true
  convert_negative_feedback_to_eval_candidate: true
  require_human_approval: true
  min_score_for_candidate: 5
```

Flow:

```text
Bad production output
        ↓
Human marks 3/10, wrong answer
        ↓
AgentVoir stores feedback
        ↓
Trace becomes eval candidate
        ↓
Reviewer approves
        ↓
New regression test added
        ↓
Future agent versions must pass it
```

This is exactly how AgentVoir can become more than a passive registry.

---

# 9. Suggested database entities

You should add these tables/entities:

```text
agent_feedback
agent_quality_scores
agent_review_jobs
agent_review_assignments
agent_eval_candidates
agent_quality_gate_events
agent_reviewer_profiles
```

Example core table:

```sql
CREATE TABLE agent_feedback (
    id UUID PRIMARY KEY,
    agent_id TEXT NOT NULL,
    agent_version TEXT,
    run_id TEXT NOT NULL,
    trace_id TEXT,
    reviewer_type TEXT NOT NULL, -- human, agent, automated_eval
    reviewer_id TEXT,
    score INTEGER CHECK (score BETWEEN 1 AND 10),
    accuracy_score INTEGER,
    usefulness_score INTEGER,
    completeness_score INTEGER,
    grounding_score INTEGER,
    safety_score INTEGER,
    format_score INTEGER,
    tool_use_score INTEGER,
    category TEXT,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

Aggregated score table:

```sql
CREATE TABLE agent_quality_scores (
    agent_id TEXT PRIMARY KEY,
    current_score NUMERIC,
    seven_day_avg NUMERIC,
    thirty_day_avg NUMERIC,
    ninety_day_avg NUMERIC,
    human_score_avg NUMERIC,
    agent_score_avg NUMERIC,
    feedback_count_30d INTEGER,
    negative_feedback_rate NUMERIC,
    hallucination_rate NUMERIC,
    policy_violation_rate NUMERIC,
    last_updated_at TIMESTAMP NOT NULL DEFAULT now()
);
```

---

# 10. Add this to AgentVoir’s metadata model

The new section could look like this:

```yaml
quality_feedback:
  enabled: true

  score:
    current: 8.4
    seven_day_avg: 8.6
    thirty_day_avg: 8.4
    confidence: 0.78

  dimensions:
    accuracy: 8.7
    usefulness: 8.2
    completeness: 7.9
    grounding: 8.8
    safety: 9.5
    format_correctness: 8.1
    tool_use_correctness: 8.6

  sampling:
    strategy: "risk_weighted"
    human_review_frequency: "daily"
    agent_review_frequency: "daily"

  gates:
    min_quality_score: 7.0
    min_safety_score: 9.0
    action_on_breach: "require_human_review"

  feedback_to_eval:
    enabled: true
    convert_negative_feedback: true
```

---

## My recommendation

Yes, AgentVoir should capture quality feedback.

But I would implement it as **three layers**, not just one 1–10 rating:

```text
Layer 1: User feedback
Simple thumbs up/down, 1–10 score, optional comment.

Layer 2: Human reviewer feedback
Structured daily reviews for important/high-risk agents.

Layer 3: Agent reviewer feedback
Automated daily review by evaluator agents for grounding, safety, format, policy, and tool use.
```

Then connect it to:

```text
quality dashboards
risk scores
runtime gates
human approval
eval datasets
agent quarantine
model routing decisions
```

This would be a very strong AgentVoir feature because commercial tools talk a lot about observability and evals, but an **agent reputation system** with human + agent feedback could become a distinctive part of your product.