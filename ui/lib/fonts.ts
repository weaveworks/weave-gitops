import { createGlobalStyle } from "styled-components";
// Typescript will handle type-checking/linting for this file
// @ts-expect-error TODO
import ProximaNovaBold from "url:../fonts/ProximaNovaBold.otf";
// @ts-expect-error TODO
import ProximaNovaRegular from "url:../fonts/ProximaNovaRegular.otf";
// @ts-expect-error TODO
import ProximaNovaSemibold from "url:../fonts/ProximaNovaSemibold.otf";
// @ts-expect-error TODO
import robotomono from "url:../fonts/roboto-mono-regular.woff";

const Fonts = createGlobalStyle`
    @font-face {
        font-family: 'proxima-nova';
        src: url(${ProximaNovaBold});
        font-weight: bold;
    }
    @font-face {
        font-family: 'proxima-nova';
        src: url(${ProximaNovaSemibold});
        font-weight: 600;
    }
    @font-face {
        font-family: 'proxima-nova';
        src: url(${ProximaNovaRegular});
        font-weight: 400;
    }
    @font-face {
        font-family: 'Roboto Mono';
        src: url(${robotomono});
    }
`;

export default Fonts;
