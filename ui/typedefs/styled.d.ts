/* eslint-disable */
import "styled-components";
export namespace colors {
  const black: string;
  const white: string;
  const primary: string;
  const primaryLight: string;
  const primaryDark: string;
  const success: string;
  const alert: string;
  const neutral00: string;
  const neutral10: string;
  const neutral20: string;
  const neutral30: string;
  const neutral40: string;
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
  const normal: string;
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
  }
}
