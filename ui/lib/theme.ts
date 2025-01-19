// Typescript will handle type-checking/linting for this file
import { createTheme, alpha } from "@mui/material/styles";
// eslint-disable-next-line
import { createGlobalStyle, DefaultTheme } from "styled-components";
import { ThemeTypes } from "../contexts/AppContext";
import images from "./images";

const baseSpacingNumber = 16;

export const baseTheme = {
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

export const theme = (mode: ThemeTypes = ThemeTypes.Light): DefaultTheme => {
  //dark
  if (mode === ThemeTypes.Dark)
    return {
      ...baseTheme,
      colors: {
        black: "#1a1a1a",
        white: "#ffffff",
        primary: "#009CCC",
        //only used in nav text when collapsed + selected/hover
        primaryLight05: "rgba(0,179,236,0.05)",
        primaryLight10: "#98E0F7",
        primary10: "#00b3ec",
        primary20: "#006B8E",
        primary30: "#556991",
        successLight: "#156034",
        successMedium: "#78CC9C",
        successOriginal: "#27AE60",
        successDark: "#C9EBD7",
        alertLight: "#9F3119",
        alertMedium: "#D58572",
        alertOriginal: "#BC3B1D",
        alertDark: "#9F3119",
        neutralGray: "#32324B",
        pipelineGray: "#4b5778",
        neutral00: "#1a1a1a",
        neutral10: "#737373",
        neutral20: "#d8d8d8",
        neutral30: "#f5f5f5",
        neutral40: "#ffffff",
        whiteToPrimary: "#32324B",
        grayToPrimary: "#009CCC",
        neutral20ToPrimary: "#32324B",
        backGray: "#32324B",
        pipelinesBackGray: "#4b5778",
        blueWithOpacity: "rgba(0, 179, 236, 0.1)",
        feedbackLight: "#8A460A",
        feedbackMedium: "#F7BF8E",
        feedbackOriginal: "#F2994A",
        feedbackDark: "#FCE6D2",
        defaultLight: "#FCE6D2",
        defaultMedium: "#F7BF8E",
        defaultOriginal: "#F2994A",
        defaultDark: "#8A460A",
      },
      mode: ThemeTypes.Dark,
    };
  //light
  else
    return {
      ...baseTheme,
      colors: {
        black: "#1a1a1a",
        white: "#ffffff",
        primary: "#00b3ec",
        primaryLight05: "#E5F7FD",
        primaryLight10: "#98E0F7",
        primary10: "#009CCC",
        primary20: "#006B8E",
        primary30: "#556991",
        successLight: "#C9EBD7",
        successMedium: "#78CC9C",
        successOriginal: "#27AE60",
        successDark: "#156034",
        alertLight: "#EECEC7",
        alertMedium: "#D58572",
        alertOriginal: "#BC3B1D",
        alertDark: "#9F3119",
        neutralGray: "#C2C9D7",
        pipelineGray: "#dde1e9",
        neutral00: "#EEEEEE",
        neutral10: "#f5f5f5",
        neutral20: "#BDBDBD",
        neutral30: "#737373",
        neutral40: "#1a1a1a",
        whiteToPrimary: "#fff",
        grayToPrimary: "#737373",
        neutral20ToPrimary: "#d8d8d8",
        backGray: "#eef0f4",
        pipelinesBackGray: "#eef0f4",
        blueWithOpacity: "rgba(0, 179, 236, 0.1)",
        feedbackLight: "#FCE6D2",
        feedbackMedium: "#F7BF8E",
        feedbackOriginal: "#F2994A",
        feedbackDark: "#8A460A",
        defaultLight: "#FCE6D2",
        defaultMedium: "#F7BF8E",
        defaultOriginal: "#F2994A",
        defaultDark: "#8A460A",
      },
      mode: ThemeTypes.Light,
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
    color: ${(props) => props.theme.colors.neutral40};
    padding: 0;
    margin: 0;
    min-width: fit-content;
    background: right bottom no-repeat fixed;
    background-image: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? `url(${images.bgDark})`
        : `url(${images.bg}), linear-gradient(to bottom, rgba(85, 105, 145, .35) 5%, rgba(85, 105, 145, .30), rgba(85, 105, 145, .25) 35%)`};
    background-color: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.neutralGray
        : "transparent"};
    background-size: 100%;
  }
  .auth-modal-size {
    min-height: 475px
  }

//prevents white autofill background in dark mode
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus {
    ${(props) =>
      props.theme.mode === ThemeTypes.Dark &&
      `background-color: ${props.theme.colors.blueWithOpacity};`}
  }
`;

export const muiTheme = (colors, mode) =>
  createTheme({
    typography: { fontFamily: "proxima-nova" },
    mixins: {},
    palette: {
      mode: "light",
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
    components: {
      MuiSlider: {
        styleOverrides: {
          root: {
            color: colors.primary,
          },
        },
      },
      MuiTooltip: {
        styleOverrides: {
          tooltip: {
            fontSize: "1rem",
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            overflowX: "hidden",
            backgroundColor: colors.white,
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            width: "60%",
            minWidth: 600,
            backgroundColor: colors.neutral00,
          },
        },
      },
      //for dark mode disabled buttons
      MuiButton: {
        styleOverrides: {
          root: {
            "&.Mui-disabled": {
              color:
                mode === ThemeTypes.Dark ? colors.primary30 : colors.neutral20,
            },
          },
          outlined: {
            "&.Mui-disabled": {
              borderColor:
                mode === ThemeTypes.Dark ? colors.primary30 : colors.neutral20,
            },
            "&.Mui-outlinedPrimary": {
              borderColor:
                mode === ThemeTypes.Dark ? colors.primary30 : colors.neutral20,
            },
          },
        },
      },
      //disabled checkboxes in dark mode
      MuiCheckbox: {
        styleOverrides: {
          root: {
            "&.Mui-disabled": {
              color: mode === ThemeTypes.Dark && colors.neutral40,
            },
          },
        },
      },
      MuiInput: {
        styleOverrides: {
          underline: {
            "&::before": {
              borderBottom:
                mode === ThemeTypes.Dark && `1px solid ${colors.neutral40}`,
            },
          },
        },
      },
      // radio buttons
      MuiRadio: {
        styleOverrides: {
          root: {
            padding: 0,
            color: colors.primary30,

            "&:hover": {
              backgroundColor: ThemeTypes.Dark
                ? alpha(colors.primary10, 0.2)
                : alpha(colors.primary, 0.1),
              color: colors.primary10,
            },

            "&.Mui-checked": {
              color: colors.primary10,
            },

            "&.Mui-disabled": {
              color:
                mode === ThemeTypes.Dark ? colors.primary30 : colors.neutral20,
            },
          },
        },
      },
    },
  });
