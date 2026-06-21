import { ComingSoon } from "../components/ComingSoon";

export const metadata = { title: "Traces | AgentVoir" };

export default function TracesPage() {
  return (
    <ComingSoon
      title="Traces"
      description="Query mode with custom time ranges and server-side pagination for historical investigation."
      docHref="/docs/architecture/ui-dashboard.md"
    />
  );
}
