import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import Logo from "../Logo";

describe("Logo", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer.create(withTheme(<Logo />)).toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
