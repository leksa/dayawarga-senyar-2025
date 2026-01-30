/**
 * Creating a sidebar enables you to:
 - create an ordered group of related docs
 - render informative sidebar for each doc page
 - provide curating in form of next/prev navigation, and even categories for pages

 * See: https://docusaurus.io/docs/sidebar

 * If you want to manage sidebar manually, remove it from here
 */

module.exports = {
  docs: {
    Introduction: ['introduction'],
    Locations: [
      'locations/overview',
      'locations/list-locations',
      'locations/get-location-by-id',
    ],
  },
};