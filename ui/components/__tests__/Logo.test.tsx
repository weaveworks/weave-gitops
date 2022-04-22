import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withContext, withTheme } from "../../lib/test-utils";
import Logo from "../Logo";

describe("Logo", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(withTheme(withContext(<Logo />, "", {})))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
