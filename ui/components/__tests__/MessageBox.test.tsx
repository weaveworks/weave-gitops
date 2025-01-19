import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import MessageBox from "../MessageBox";

describe("MessageBox", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = render(
        withTheme(<MessageBox align>Column and items centered.</MessageBox>),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
