// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightLlmsTxt from 'starlight-llms-txt';

export default defineConfig({
  site: 'https://cli.stocksmith.dev',
  integrations: [
    starlight({
      plugins: [starlightLlmsTxt()],
      title: 'Stocksmith CLI',
      description: 'The command-line interface for Stocksmith.',
      favicon: '/favicon.svg',
      // Logos are named by their own colour, so map by theme function:
      // black wordmark on light backgrounds, white wordmark on dark backgrounds.
      logo: {
        light: './src/assets/logo-dark.svg',
        dark: './src/assets/logo-light.svg',
        replacesTitle: true,
        alt: 'Stocksmith CLI',
      },
      head: [
        { tag: 'meta', attrs: { property: 'og:image', content: 'https://cli.stocksmith.dev/favicon.svg' } },
        { tag: 'meta', attrs: { name: 'theme-color', content: '#3EB1C1' } },
      ],
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/craftybase/stocksmith-cli' },
      ],
      sidebar: [
        { label: 'Getting Started', slug: 'getting-started' },
        { label: 'Updating', slug: 'updating' },
        { label: 'Authentication', slug: 'authentication' },
        { label: 'Output Formats', slug: 'output-formats' },
        { label: 'Configuration', slug: 'configuration' },
        { label: 'Pagination', slug: 'pagination' },
        {
          label: 'Command Reference',
          items: [{ autogenerate: { directory: 'reference' } }],
        },
        { label: 'Using with Agents & LLMs', slug: 'agents' },
      ],
    }),
  ],
});
