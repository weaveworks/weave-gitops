// {
//   "/v1": {
//     "target": "https://localhost:9001/",
//   },
//   "/oauth2": {
//     "target": "https://localhost:9001/"
//   },
// }

const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  app.use(
    createProxyMiddleware("/v1", {
      target: "https://localhost:9001/",
    })
  );
};
