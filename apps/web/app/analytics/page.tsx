import { ComingSoon } from "../components/ComingSoon";

export const metadata = { title: "Analytics | AgentVoir" };

export default function AnalyticsPage() {
  return (
    <ComingSoon
      title="Analytics"
      description="Aggregated usage, cost, and quality trends — not the raw live event stream."
      docHref="/docs/architecture/ui-dashboard.md"
    />
  );
}
