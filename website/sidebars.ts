import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";
import config from "./docusaurus.config";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)
const sidebars: SidebarsConfig = {
  manualSidebar: (config.customFields?.pragmaticSidebar as SidebarsConfig[string]) ?? [],
};

export default sidebars;
