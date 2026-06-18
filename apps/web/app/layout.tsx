import type { Metadata } from "next";
import "./styles.css";

export const metadata: Metadata = {
  title: "AgentVoir",
  description: "Enterprise AI agent registry and LLM gateway"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
