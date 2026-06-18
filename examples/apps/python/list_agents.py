from agentvoir import AgentVoirClient

client = AgentVoirClient(base_url="http://localhost:8081", api_key="dev")
print(client.health())
print(client.list_agents())
