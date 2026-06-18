# AgentVoir evaluator

The evaluator service runs agent evals and regression checks against datasets for the AgentVoir control plane. It is intended to validate agent behavior, catch regressions, and compare runs over time as prompts, models, and policies change. The repository layout reserves this module as the agent eval runner alongside the gateway, registry, and other services. Implementation is still a scaffold: the service structure is in place, but eval runners, dataset integration, and reporting are placeholders until the first implementation milestone.
