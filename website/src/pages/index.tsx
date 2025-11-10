import type { ReactNode } from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import HomepageFeatures from "@site/src/components/HomepageFeatures";
import Heading from "@theme/Heading";

import styles from "./index.module.css";

function HomepageHeader() {
    const { siteConfig } = useDocusaurusContext();
    return (
        <header className={clsx("hero", styles.heroBanner)}>
            <div className="container">
                <Heading
                    as="h1"
                    className={clsx("hero__title", styles.heroTitle)}
                >
                    {siteConfig.title}
                </Heading>
                <p className={clsx("hero__subtitle", styles.heroSubtitle)}>
                    {siteConfig.tagline}
                </p>
                <div className={styles.buttons}>
                    <Link
                        className="button button--info button--lg"
                        to="/docs/quickstart"
                    >
                        Get Started
                    </Link>
                </div>
            </div>
        </header>
    );
}

export default function Home(): ReactNode {
    const { siteConfig } = useDocusaurusContext();
    return (
        <Layout
            title="Terminal-based Personal Information Manager"
            description="Manage tasks, notes, articles, and media from your terminal with Noteleaf"
        >
            <HomepageHeader />
            <main>
                <HomepageFeatures />
            </main>
        </Layout>
    );
}
