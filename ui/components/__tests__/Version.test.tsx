import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import Version from "../Version";

const productName = "product name";
const versionText = "version text";
const versionHref = "https://github.com/weaveworks/weave-gitops";

describe("Version", () => {
  describe("snapshots", () => {
    it("renders plain text with version text and without version href", () => {
      const tree = render(
        withTheme(
          <Version productName={productName} appVersion={{ versionText }} />,
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("renders a link with version text and version href", () => {
      const tree = render(
        withTheme(
          <Version
            productName={productName}
            appVersion={{ versionText, versionHref }}
          />,
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("renders a dash without version text and without version href", () => {
      const tree = render(
        withTheme(
          <Version
            productName={productName}
            appVersion={{ versionText: "" }}
          />,
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("renders a dash as plain text without version text and with version href", () => {
      const tree = render(
        withTheme(
          <Version
            productName={productName}
            appVersion={{ versionText: "", versionHref }}
          />,
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
