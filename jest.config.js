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
  transformIgnorePatterns: ["/node_modules/(?!(yaml)/)"],
  setupFilesAfterEnv: ["<rootDir>/setup-jest.ts"],
  modulePathIgnorePatterns: ["<rootDir>/dist/"],
  testEnvironment: "jsdom",
};

export default config;
