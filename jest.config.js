const transformIgnorePatterns = [
  '@emotion\/.*',
  '@mui\/.*',
  'd3',
  'd3-dag',
  'history',
  'http-proxy-middleware',
  'install',
  'jest-canvas-mock',
  'js-sha3',
  'lodash',
  'luxon',
  'mnemonic-browser',
  'postcss',
  'react',
  'react-.*',
  'remark-gfm',
  'styled-components',
].join('|');

/** @type {import('jest').Config} */
const config = {
  preset: "ts-jest",
  moduleNameMapper: {
    "\\.(jpg|ico|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$":
      "<rootDir>/ui/lib/fileMock.js",
    "\\.(css|less)$": "<rootDir>/ui/lib/fileMock.js",
  },
  transform: {
    "\\.tsx?$": "ts-jest",
    "\\.jsx?$": [
      "babel-jest",
      {
        configFile: "./babel.config.json",
      },
    ],
  },
  transformIgnorePatterns: [`/node_modules/(?:${transformIgnorePatterns})/`],
  setupFilesAfterEnv: ["<rootDir>/setup-jest.ts"],
  modulePathIgnorePatterns: ["<rootDir>/dist/"],
  testEnvironment: "jsdom",
};

export default config;
