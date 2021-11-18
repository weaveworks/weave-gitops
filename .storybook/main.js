// module.exports = {
//   stories: [
//     "../ui/stories/**/*.stories.mdx",
//     "../ui/stories/**/*.stories.@(js|jsx|ts|tsx)",
//   ],
//   addons: ["@storybook/addon-links", "@storybook/addon-essentials"],
// };
const path = require("path");

// Export a function. Accept the base config as the only param.
module.exports = {
  stories: [
    "../ui/stories/**/*.stories.mdx",
    "../ui/stories/**/*.stories.@(js|jsx|ts|tsx)",
  ],
  addons: ["@storybook/addon-links", "@storybook/addon-essentials"],
  webpackFinal: async (config, { configType }) => {
    // `configType` has a value of 'DEVELOPMENT' or 'PRODUCTION'
    // You can change the configuration based on that.
    // 'PRODUCTION' is used when building the static version of storybook.

    // Make whatever fine-grained changes you need
    config.module.rules.push({
      test: /url:/i,
      use: [
        {
          loader: "url-loader",
          options: {
            limit: 8192,
            generator: (content, mimetype, encoding, resourcePath) => {
              console.log("Josh " + resourcePath);
              return content;
            },
          },
        },
      ],
      include: path.resolve(__dirname, "../"),
    });
    // Return the altered config
    return config;
  },
};

// module.exports = {
//   module: {
//     rules: [
//       {
//         test: /\.(png|jpg|gif)$/i,
//         use: [
//           {
//             loader: 'url-loader',
//             options: {
//               limit: 8192,
//             },
//           },
//         ],
//       },
//     ],
//   },
// };
