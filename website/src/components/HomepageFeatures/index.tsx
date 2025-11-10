import Heading from "@theme/Heading";
import clsx from "clsx";
import type { ReactNode } from "react";
import styles from "./styles.module.css";

type FeatureItem = { title: string; description: ReactNode; theme: keyof typeof themeClassMap };

const themeClassMap = {
  primary: styles.featureCardPrimary,
  info: styles.featureCardInfo,
  success: styles.featureCardSuccess,
  warning: styles.featureCardWarning,
  accent: styles.featureCardAccent,
  plum: styles.featureCardPlum,
  danger: styles.featureCardDanger,
};

const FeatureList: FeatureItem[] = [{
  title: "Terminal-First Interface",
  description: (
    <>Built for the command line with a beautiful, keyboard-driven interface powered by Bubble Tea and Fang.</>
  ),
  theme: "primary",
}, {
  title: "Task Management",
  description: (
    <>Organize your tasks with projects, priorities, tags, contexts, due dates, and recurrence—all from a single CLI.</>
  ),
  theme: "danger",
}, {
  title: "Leaflet.pub Publishing",
  description: (
    <>
      Sync Markdown notes with leaflet.pub, push updates over AT Protocol, and manage drafts without leaving the
      terminal.
    </>
  ),
  theme: "accent",
}, {
  title: "Articles & Readability",
  description: (
    <>
      Capture the clean content of any article, store Markdown + HTML copies, and enjoy a terminal reader inspired by
      Readability.
    </>
  ),
  theme: "warning",
}, {
  title: "Knowledge Base",
  description: (
    <>Keep notes, track books, movies, and TV shows. Link everything together with tags, IDs, and shared metadata.</>
  ),
  theme: "info",
}, {
  title: "Open Source & MIT Licensed",
  description: (
    <>
      Built in the open on GitHub under the MIT license. Fork it, extend it, and make Noteleaf part of your own
      workflows.
    </>
  ),
  theme: "plum",
}];

function Feature({ title, description, theme }: FeatureItem) {
  const cardClass = themeClassMap[theme] ?? themeClassMap.primary;
  return (
    <div className={clsx("col col--4")}>
      <div className={clsx("text--left padding-horiz--md", styles.featureCard, cardClass)}>
        <Heading as="h3" className={styles.featureTitle}>{title}</Heading>
        <p className={styles.featureCopy}>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">{FeatureList.map((props, idx) => <Feature key={idx} {...props} />)}</div>
      </div>
    </section>
  );
}
