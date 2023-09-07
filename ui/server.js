const fs = require("fs");
const express = require("express");
const { createProxyMiddleware } = require("http-proxy-middleware");

const proxyHost = process.env.PROXY_HOST || "http://localhost:9001/";
// Accept self-signed certificates etc (for local development).
// If you are using demo-01 etc you can set PROXY_SECURE=true and it should work.
const secure = process.env.PROXY_SECURE === "true";

const subPath = process.env.WEGO_SUB_PATH || "/foo";
const proxyMiddleWare = createProxyMiddleware({
  target: proxyHost,
  changeOrigin: true,
  pathRewrite: {
    [`^${subPath}`]: "/", // remove base path
  },

  secure,
});

const renderIndexHtml = (req, res, next) => {
  const writeIndexResponse = (err, result) => {
    if (err) {
      return next(err);
    }
    res.set("content-type", "text/html");
    // read result into a string
    const resultString = result.toString();
    // replace "<head>" with "<head><base href="/foo/">"
    const resultStringWithBaseHref = resultString.replace(
      "<head>",
      `<head><base href='${subPath}/' />`
    );

    // send the modified result
    res.send(resultStringWithBaseHref);
  };
  fs.readFile("bin/dist/index.html", writeIndexResponse);
};

const appRouter = express.Router();

appRouter.get("", renderIndexHtml);
// static files
appRouter.use(express.static("bin/dist"));

// serve index.html on react-router's browserHistory paths
// LIST OUT PATHS EXPLICITLY SO PROXY_HOST WILL STILL WORK.
//
appRouter.use(
  [
    "/clusters",
    "/templates",
    "/applications",
    "/application_add",
    "/application_detail",
    "/application_remove",
    "/sign_in",
  ],
  renderIndexHtml
);

appRouter.use(["/v1", "/debug", "/oauth2"], proxyMiddleWare);

var expressApp = express();
expressApp.use(subPath, appRouter);

const port = process.env.PORT || 5001;
const server = expressApp.listen(port, () => {
  let { address } = server.address();
  if (address.indexOf(":") !== -1) {
    address = `[${address}]`;
  }
  console.log("weave-gitops listening at http://%s:%s", address, port);
});
