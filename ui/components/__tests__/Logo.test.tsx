import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withContext, withTheme } from "../../lib/test-utils";
import Logo from "../Logo";

describe("Logo", () => {
  describe("snapshots", () => {
    it("renders open view", () => {
      const tree = renderer
        .create(withTheme(withContext(<Logo collapsed={false} />, "", {})))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("renders collapsed view", () => {
      const tree = renderer
        .create(withTheme(withContext(<Logo collapsed={true} />, "", {})))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
