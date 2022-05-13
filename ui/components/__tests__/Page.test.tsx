import "jest-styled-components";
import "jest-canvas-mock";
import React from "react";
import renderer from "react-test-renderer";
import { withContext, withTheme } from "../../lib/test-utils";
import Page from "../Page";

describe("Page", () => {
  describe("snapshots", () => {
    it("default", () => {
      const tree = renderer
        .create(withTheme(withContext(<Page />, "", {})))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
