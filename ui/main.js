import { createBrowserHistory } from "history";
import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";

const history = createBrowserHistory();

// eslint-disable-next-line
const container = document.getElementById("app");
const root = createRoot(container);
root.render(<App history={history} />);
// eslint-disable-next-line
if (module.hot) {
  // eslint-disable-next-line
  module.hot.accept();
}
