// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'DayaWarga API Documentation',
  tagline: 'Platform manajemen data posko pengungsi, fasilitas kesehatan, dan infrastruktur',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://dayawarga.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  baseUrl: '/api-docs/',

  // GitHub pages deployment config
  organizationName: 'dayawarga',
  projectName: 'senyar-2025',

  onBrokenLinks: 'warn',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          routeBasePath: '/',
          // Please change this to your repo.
          editUrl: 'https://github.com/dayawarga/senyar-2025/edit/main/api-docs/docs/',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/docusaurus-social-card.jpg',
      navbar: {
        title: 'DayaWarga API',
        logo: {
          alt: 'DayaWarga Logo',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docs',
            position: 'left',
            label: 'API Reference',
          },
          {
            href: 'https://github.com/dayawarga/senyar-2025',
            label: 'GitHub',
            position: 'right',
          },
          {
            href: 'https://dayawarga.com',
            label: 'Website',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Introduction',
                to: '/introduction',
              },
              {
                label: 'Locations API',
                to: '/locations/overview',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/dayawarga/senyar-2025',
              },
              {
                label: 'Website',
                href: 'https://dayawarga.com',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/dayawarga/senyar-2025',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} DayaWarga. Built with Docusaurus.`,
      },
    }),
};

module.exports = config;