import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const ghURL = "https://github.com/stormlightlabs/noteleaf";
// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)
const config: Config = {
    title: "My Site",
    tagline: "Dinosaurs are cool",
    favicon: "img/favicon.ico",
    // Improve compatibility with the upcoming Docusaurus v4
    future: { v4: true },
    url: "https://stormlightlabs.github.io/",
    baseUrl: "/noteleaf/",
    organizationName: "stormlightlabs",
    projectName: "noteleaf",
    onBrokenLinks: "throw",
    i18n: { defaultLocale: "en", locales: ["en"] },
    presets: [
        [
            "classic",
            {
                docs: { sidebarPath: "./sidebars.ts" },
                theme: { customCss: "./src/css/custom.css" },
            } satisfies Preset.Options,
        ],
    ],

    themeConfig: {
        image: "img/docusaurus-social-card.jpg",
        colorMode: { respectPrefersColorScheme: true },
        navbar: {
            title: "My Site",
            logo: { alt: "My Site Logo", src: "img/logo.svg" },
            items: [
                {
                    type: "docSidebar",
                    sidebarId: "manualSidebar",
                    position: "left",
                    label: "Manual",
                },
                { href: ghURL, label: "GitHub", position: "right" },
            ],
        },
        footer: {
            style: "dark",
            links: [
                {
                    title: "Docs",
                    items: [
                        {
                            label: "Tutorial",
                            to: "/docs/intro",
                        },
                    ],
                },
                {
                    title: "Community",
                    items: [
                        {
                            label: "BlueSky",
                            href: "https://bsky.app/desertthunder.dev",
                        },
                        {
                            label: "X",
                            href: "https://x.com/_desertthunder",
                        },
                    ],
                },
                {
                    title: "More",
                    items: [{ label: "GitHub", href: ghURL }],
                },
            ],
            copyright: `Copyright © ${new Date().getFullYear()} Stormlight Labs, LLC.`,
        },
        prism: {
            theme: prismThemes.github,
            darkTheme: prismThemes.dracula,
        },
    } satisfies Preset.ThemeConfig,
};

export default config;
