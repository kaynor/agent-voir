import { ComingSoon } from "../components/ComingSoon";

export const metadata = { title: "Models | AgentVoir" };

export default function ModelsPage() {
  return (
    <ComingSoon
      title="Models & providers"
      description="Provider health, rate limits, pricing drift, and model catalog."
      docHref="/docs/architecture/ui-dashboard.md"
    />
  );
}
