// Typescript will handle type-checking/linting for this file
import { createTheme } from "@material-ui/core";
// eslint-disable-next-line
import { createGlobalStyle, DefaultTheme } from "styled-components";
import images from "./images";

const baseSpacingNumber = 16;

const baseTheme = {
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
  spacing: {
    none: "0",
    // 4px
    xxs: `${baseSpacingNumber * 0.25}px`,
    // 8px
    xs: `${baseSpacingNumber * 0.5}px`,
    // 12px
    small: `${baseSpacingNumber * 0.75}px`,
    // 16px
    base: `${baseSpacingNumber}px`,
    // 24px
    medium: `${baseSpacingNumber * 1.5}px`,
    // 32px
    large: `${baseSpacingNumber * 2}px`,
    // 48px
    xl: `${baseSpacingNumber * 3}px`,
    // 64px
    xxl: `${baseSpacingNumber * 4}px`,
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

export const theme = (mode: "light" | "dark"): DefaultTheme => {
  //dark
  if (mode === "dark")
    return {
      ...baseTheme,
      colors: {
        black: "#fff",
        white: "#1a1a1a",
        primary: "#009CCC",
        primaryLight05: "#E5F7FD",
        primaryLight10: "#98E0F7",
        primary10: "#009CCC",
        primary20: "#006B8E",
        successLight: "#C9EBD7",
        successMedium: "#78CC9C",
        successOriginal: "#27AE60",
        successDark: "#156034",
        alertLight: "#EECEC7",
        alertMedium: "#D58572",
        alertOriginal: "#BC3B1D",
        alertDark: "#9F3119",
        neutralGray: "#32324B",
        neutral00: "#1a1a1a",
        neutral10: "#737373",
        neutral20: "#d8d8d8",
        neutral30: "#f5f5f5",
        neutral40: "#ffffff",
        backGrey: "#eef0f4",
        feedbackLight: "#FCE6D2",
        feedbackMedium: "#F7BF8E",
        feedbackOriginal: "#F2994A",
        feedbackDark: "#8A460A",
        defaultLight: "#FCE6D2",
        defaultMedium: "#F7BF8E",
        defaultOriginal: "#F2994A",
        defaultDark: "#8A460A",
      },
    };
  //light
  else
    return {
      ...baseTheme,
      colors: {
        black: "#1a1a1a",
        white: "#fff",
        primary: "#00b3ec",
        primaryLight05: "#E5F7FD",
        primaryLight10: "#98E0F7",
        primary10: "#009CCC",
        primary20: "#006B8E",
        successLight: "#C9EBD7",
        successMedium: "#78CC9C",
        successOriginal: "#27AE60",
        successDark: "#156034",
        alertLight: "#EECEC7",
        alertMedium: "#D58572",
        alertOriginal: "#BC3B1D",
        alertDark: "#9F3119",
        neutralGray: "#F6F7F9",
        neutral00: "#ffffff",
        neutral10: "#f5f5f5",
        neutral20: "#d8d8d8",
        neutral30: "#737373",
        neutral40: "#1a1a1a",
        backGrey: "#eef0f4",
        feedbackLight: "#FCE6D2",
        feedbackMedium: "#F7BF8E",
        feedbackOriginal: "#F2994A",
        feedbackDark: "#8A460A",
        defaultLight: "#FCE6D2",
        defaultMedium: "#F7BF8E",
        defaultOriginal: "#F2994A",
        defaultDark: "#8A460A",
      },
    };
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
  //scrollbar
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
  //MuiTabs
  .horizontal-tabs {
    .MuiTab-root {
      line-height: 1;
      letter-spacing: 1px;
      height: 32px;
      min-height: 32px;
      width: fit-content;
      @media (min-width: 600px) {
        min-width: 132px;
      }
    }
    .MuiTabs-root {
      min-height: 32px;
      margin: ${(props) => props.theme.spacing.xs} 0;
    }
    .MuiTabs-fixed {
      height: 32px;
    }
    .MuiTabs-indicator {
      height: 3px;
      background-color: ${(props) => props.theme.colors.primary};
    }
  }
`;

export const muiTheme = (colors) =>
  createTheme({
    typography: { fontFamily: "proxima-nova" },
    palette: {
      primary: {
        //Main - Primary Color Dark - 10
        main: colors.primary10,
      },
      secondary: {
        //Feedback - Alert - Original
        main: colors.alertOriginal,
      },
      text: {
        //Neutral - Neutral - 40
        primary: colors.neutral40,
        //Neutral - Neutral - 30
        secondary: colors.neutral30,
        disabled: colors.neutral30,
      },
    },
    overrides: {
      MuiSlider: {
        root: {
          color: colors.primary,
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
      MuiDrawer: {
        paper: {
          width: "60%",
          minWidth: 600,
        },
      },
    },
  });
