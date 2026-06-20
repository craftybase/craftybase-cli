// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightLlmsTxt from 'starlight-llms-txt';

export default defineConfig({
  site: 'https://cli.craftybase.dev',
  integrations: [
    starlight({
      plugins: [starlightLlmsTxt()],
      title: 'Craftybase CLI',
      description: 'The command-line interface for Craftybase.',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/craftybase/craftybase-cli' },
      ],
      sidebar: [
        { label: 'Getting Started', slug: 'getting-started' },
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
