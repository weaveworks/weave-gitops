/* eslint-disable */
import "styled-components";
import { ThemeTypes } from "../contexts/AppContext";
export namespace colors {
  const black: string;
  const white: string;
  const primary: string;
  const primaryLight05: string;
  const primaryLight10: string;
  const primary10: string;
  const primary20: string;
  const primary30: string;
  const alertLight: string;
  const alertOriginal: string;
  const alertMedium: string;
  const alertDark: string;
  const neutralGray: string;
  const neutral00: string;
  const neutral10: string;
  const neutral20: string;
  const neutral30: string;
  const neutral40: string;
  const whiteToPrimary: string;
  const grayToPrimary: string;
  const neutral20ToPrimary: string;
  const backGray: string;
  const blueWithOpacity: string;
  const feedbackLight: string;
  const feedbackMedium: string;
  const feedbackOriginal: string;
  const feedbackDark: string;
  const successLight: string;
  const successMedium: string;
  const successOriginal: string;
  const successDark: string;
  const defaultLight: string;
  const defaultMedium: string;
  const defaultOriginal: string;
  const defaultDark: string;
}

export namespace spacing {
  const base: string;
  const large: string;
  const medium: string;
  const none: string;
  const small: string;
  const xl: string;
  const xs: string;
  const xxl: string;
  const xxs: string;
}

export namespace fontSizes {
  const huge: string;
  const extraLarge: string;
  const large: string;
  const medium: string;
  const small: string;
  const tiny: string;
}

export namespace borderRadius {
  const circle: string;
  const none: string;
  const soft: string;
}

export namespace boxShadow {
  const light: string;
  const none: string;
}

declare module "styled-components" {
  export interface DefaultTheme {
    fontFamilies: { regular: string; monospace: string };
    fontSizes: typeof fontSizes;
    colors: typeof colors;
    spacing: typeof spacing;
    borderRadius: typeof borderRadius;
    boxShadow: typeof boxShadow;
    mode: ThemeTypes;
  }
}
