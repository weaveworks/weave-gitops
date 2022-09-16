// Typescript will handle type-checking/linting for this file
import { createTheme } from "@material-ui/core";
// eslint-disable-next-line
import { createGlobalStyle, DefaultTheme } from "styled-components";
import images from "./images";

const baseSpacingNumber = 16;

export const theme: DefaultTheme = {
  fontFamilies: {
    monospace: "'Roboto Mono', monospace",
    regular: "'proxima-nova', Helvetica, Arial, sans-serif",
  },
  fontSizes: {
    huge: "48px",
    extraLarge: "32px",
    large: "20px",
    medium: "14px",
    small: "12px",
    tiny: "10px",
  },
  colors: {
    black: "#1a1a1a",
    white: "#fff",
    primary: "#00b3ec",
    primaryLight: "#98E0F7",
    primary10: "#009CCC",
    primary20: "#006B8E",
    success: "#27AE60",
    alert: "#BC381D",
    suspended: "#f2994a",
    neutralGray: "#F6F7F9",
    neutral00: "#ffffff",
    neutral10: "#f5f5f5",
    neutral20: "#d8d8d8",
    neutral30: "#737373",
    neutral40: "#1a1a1a",
    backGrey: "#eef0f4",
    feedbackLight: "#FCE6D2",
  },
  spacing: {
    // 16px
    base: `${baseSpacingNumber}px`,
    // 32px
    large: `${baseSpacingNumber * 2}px`,
    // 24px
    medium: `${baseSpacingNumber * 1.5}px`,
    none: "0",
    // 12px
    small: `${baseSpacingNumber * 0.75}px`,
    // 48px
    xl: `${baseSpacingNumber * 3}px`,
    // 8px
    xs: `${baseSpacingNumber * 0.5}px`,
    // 64px
    xxl: `${baseSpacingNumber * 4}px`,
    // 4px
    xxs: `${baseSpacingNumber * 0.25}px`,
  },
  borderRadius: {
    circle: "50%",
    none: "0",
    soft: "2px",
  },
  boxShadow: {
    light: "0 1px 3px #f5f5f5, 0 1px 2px #d8d8d8",
    none: "none",
  },
};

export default theme;

export const GlobalStyle = createGlobalStyle`
  html,
  body {
    height: 100%;
    margin: 0;
  }

  #app {
    display: flex;
    flex-flow: column;
    height: 100%;
    margin: 0;
  }
  
  body {
    font-family: ${(props) => props.theme.fontFamilies.regular};
    font-size: ${(props) => props.theme.fontSizes.medium};
    color: ${(props) => props.theme.colors.black};
    padding: 0;
    margin: 0;
    min-width: fit-content;
    background: right bottom no-repeat fixed; 
    background-image: url(${
      images.bg
    }), linear-gradient(to bottom, rgba(85, 105, 145, .1) 5%, rgba(85, 105, 145, .1), rgba(85, 105, 145, .25) 35%);
    background-size: 100%;
  }
  .auth-modal-size {
    min-height: 475px
  }
  ::-webkit-scrollbar-track {
    margin-top: 5px;
    -webkit-box-shadow: transparent;
    -moz-box-shadow: transparent;
    background-color: transparent;
    border-radius: 5px;
  }

  ::-webkit-scrollbar{
    width: 5px;
    height: 5px;
    background-color: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background-color: ${(props) => props.theme.colors.neutral20};
    border-radius: 5px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background-color: ${(props) => props.theme.colors.neutral30};
  }
`;

export const muiTheme = createTheme({
  typography: { fontFamily: "proxima-nova" },
  palette: {
    primary: {
      //Main - Primary Color Dark - 10
      main: theme.colors.primary10,
    },
    secondary: {
      //Feedback - Alert - Original
      main: theme.colors.alert,
    },
    text: {
      //Neutral - Neutral - 40
      primary: theme.colors.neutral40,
      //Neutral - Neutral - 30
      secondary: theme.colors.neutral30,
      disabled: theme.colors.neutral30,
    },
  },
  overrides: {
    MuiSlider: {
      root: {
        color: theme.colors.primary,
      },
    },
    MuiTooltip: {
      tooltip: {
        fontSize: "1rem",
      },
    },
    MuiPaper: {
      root: {
        overflowX: "hidden",
      },
    },
  },
});
