export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
    exclude: [
      "as",
      "theme",
      "forwardedAs",
      //from MUI ButtonProps
      "ref",
      "children",
      "tabIndex",
      "disableFocusRipple",
      "href",
      "size",
      "action",
      "centerRipple",
      "disableRipple",
      "disableTouchRipple",
      "focusRipple",
      "focusVisibleClassName",
      "onFocusVisible",
      "TouchRippleProps",
    ],
  },
  previewTabs: { "storybook/docs/panel": { index: -1 } },
};

import { MuiThemeProvider } from "@material-ui/core";
import React from "react";
import { ThemeProvider } from "styled-components";
import theme, { muiTheme } from "../ui/lib/theme";

// Storybook does not play nice with parcel 'url:..' import statements.
// We resort to some ancient front-end sorcery here to get fonts to load so components look right.
// The entire ui/ directory is exposed to storybook as static files,
// which is how it is able to request fonts this way.
const Style = () => {
  React.useEffect(() => {
    const node = document.createElement("style");
    node.innerHTML = `
      @font-face {
        font-family: 'proxima-nova';
        src: url('/fonts/ProximaNovaRegular.otf');
      }
    
      @font-face {
        font-family: 'proxima-nova';
        src: url('/fonts/ProximaNovaBold.otf');
        font-weight: bold;
      }

      @font-face {
        font-family: 'Roboto Mono';
        src: url('/fonts/roboto-mono-regular.woff');
      }
    `;
    const head = document.querySelector("head");

    head.appendChild(node);
  }, []);

  return null;
};

export const decorators = [
  (Story) => (
    <>
      <Style />
      <MuiThemeProvider theme={muiTheme}>
        <ThemeProvider theme={theme}>
          <Story />
        </ThemeProvider>
      </MuiThemeProvider>
    </>
  ),
];
