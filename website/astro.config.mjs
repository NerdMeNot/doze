// The doze documentation site. Served at doze.nerdmenot.in (the registry stays
// at /registry, proxied by a Pages Function in the registry project — see
// docs/design/website.md in the repo root).
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://doze.nerdmenot.in",
  integrations: [
    starlight({
      title: "doze",
      description:
        "Real databases on your laptop — asleep until you need them. No Docker, no JVM, no always-on stack.",
      logo: { src: "./src/assets/logo.svg", alt: "doze" },
      favicon: "/favicon.svg",
      customCss: ["./src/styles/theme.css"],
      social: [
        { icon: "github", label: "GitHub", href: "https://github.com/doze-dev/doze" },
      ],
      editLink: {
        baseUrl: "https://github.com/doze-dev/doze/edit/main/website/",
      },
      lastUpdated: true,
      head: [
        {
          tag: "link",
          attrs: { rel: "preconnect", href: "https://fonts.googleapis.com" },
        },
        {
          tag: "link",
          attrs: {
            rel: "preconnect",
            href: "https://fonts.gstatic.com",
            crossorigin: true,
          },
        },
        {
          tag: "link",
          attrs: {
            rel: "stylesheet",
            href: "https://fonts.googleapis.com/css2?family=Instrument+Serif&family=Hanken+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap",
          },
        },
      ],
      sidebar: [
        {
          label: "Why doze",
          items: [
            { slug: "why/doze" },
            { slug: "why/not-containers" },
            { slug: "why/hcl" },
            { slug: "why/alternatives" },
            { slug: "why/trust" },
          ],
        },
        {
          label: "Getting started",
          items: [
            { slug: "start/install" },
            { slug: "start/getting-started" },
            { slug: "start/concepts" },
            { slug: "start/from-docker-compose" },
          ],
        },
        {
          label: "Guides",
          items: [
            { slug: "guides/engines" },
            {
              label: "Engine recipes",
              collapsed: true,
              items: [
                { slug: "guides/recipes/postgres" },
                { slug: "guides/recipes/valkey-kvrocks" },
                { slug: "guides/recipes/documentdb" },
                { slug: "guides/recipes/s3" },
                { slug: "guides/recipes/sqs" },
                { slug: "guides/recipes/sns" },
                { slug: "guides/recipes/stacks" },
                { slug: "guides/recipes/config-layout" },
              ],
            },
            { slug: "guides/workflows" },
            { slug: "guides/modules" },
            { slug: "guides/teams-ci" },
            { slug: "guides/files-and-storage" },
            { slug: "guides/resource-footprint" },
            { slug: "guides/troubleshooting" },
            { slug: "guides/faq" },
          ],
        },
        {
          label: "Building modules",
          items: [
            { slug: "modules/overview" },
            { slug: "modules/first-module" },
            { slug: "modules/real-engines" },
            { slug: "modules/describe" },
            { slug: "modules/testing" },
            { slug: "modules/releasing" },
            { slug: "modules/publishing" },
          ],
        },
        {
          label: "Running a registry",
          items: [
            { slug: "registry/trust-architecture" },
            { slug: "registry/self-host" },
            { slug: "registry/mirror-binaries" },
            { slug: "registry/operations" },
            { slug: "registry/roadmap-hosts" },
          ],
        },
        {
          label: "Reference",
          items: [
            { slug: "reference/cli" },
            { slug: "reference/configuration" },
            { slug: "reference/lockfile" },
            { slug: "reference/environment" },
            { slug: "reference/module-index" },
            { slug: "reference/binaries" },
            { slug: "reference/extensions" },
            { slug: "reference/architecture" },
          ],
        },
      ],
    }),
  ],
});
