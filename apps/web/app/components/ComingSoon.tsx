type ComingSoonProps = {
  title: string;
  description: string;
  docHref?: string;
};

export function ComingSoon({ title, description, docHref }: ComingSoonProps) {
  return (
    <section className="coming-soon">
      <p className="eyebrow">Coming soon</p>
      <h1>{title}</h1>
      <p className="muted">{description}</p>
      {docHref ? (
        <p>
          See{" "}
          <a href={docHref} target="_blank" rel="noreferrer">
            architecture spec
          </a>{" "}
          for the full design.
        </p>
      ) : null}
    </section>
  );
}
