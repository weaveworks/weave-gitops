/* eslint-disable */
import "styled-components";
export namespace colors {
  const black: string;
  const white: string;
  const primary: string;
  const success: string;
  const neutral20: string;
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

declare module "styled-components" {
  export interface DefaultTheme {
    fontFamilies: { regular: string; monospace: string };
    fontSizes: typeof fontSizes;
    colors: typeof colors;
    spacing: typeof spacing;
  }
}
