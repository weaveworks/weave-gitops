import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { withContext, withTheme } from "../../lib/test-utils";
import Logo from "../Logo";

describe("Logo", () => {
  describe("snapshots", () => {
    it("renders open view", () => {
      const tree = render(
        withTheme(withContext(<Logo collapsed={false} />, "", {})),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("renders collapsed view", () => {
      const tree = render(
        withTheme(withContext(<Logo collapsed={true} />, "", {})),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
