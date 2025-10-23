const transformIgnorePatterns = [
  "@emotion\/.*",
  "@mui\/.*",
  "d3",
  "d3-dag",
  "history",
  "http-proxy-middleware",
  "install",
  "jest-canvas-mock",
  "js-sha3",
  "js-yaml",
  "lodash",
  "luxon",
  "mnemonic-browser",
  "postcss",
  "react",
  "react-.*",
  "remark-gfm",
  "styled-components",
].join("|");

/** @type {import('jest').Config} */
const config = {
  preset: "ts-jest",
  moduleNameMapper: {
    "\\.(jpg|ico|jpeg|png|gif|eot|otf|webp|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$":
      "<rootDir>/ui/lib/fileMock.js",
    "\\.(css|less)$": "<rootDir>/ui/lib/fileMock.js",
    "^.+\\.svg$": "jest-transformer-svg",
  },

  transform: {
    "\\.tsx?$": "ts-jest",
    "\\.jsx?$": [
      "babel-jest",
      {
        configFile: "./babel.config.testing.json",
      },
    ],
  },
  transformIgnorePatterns: [`/node_modules/(?:${transformIgnorePatterns})/`],
  setupFilesAfterEnv: ["<rootDir>/setup-jest.ts"],
  modulePathIgnorePatterns: ["<rootDir>/dist/"],
  testEnvironment: "jsdom",
  testEnvironmentOptions: {
    url: "http://localhost",
  },
};

export default config;
