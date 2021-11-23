const path = require("path");

// Export a function. Accept the base config as the only param.
module.exports = {
  stories: [
    "../ui/stories/**/*.stories.mdx",
    "../ui/stories/**/*.stories.@(js|jsx|ts|tsx)",
  ],
  addons: ["@storybook/addon-links", "@storybook/addon-essentials"],
};
