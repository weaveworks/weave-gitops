import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import Text from "../Text";

describe("Text", () => {
  describe("snapshots", () => {
    it("normal", () => {
      const tree = renderer.create(withTheme(<Text>some text</Text>)).toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("bold", () => {
      const tree = renderer
        .create(withTheme(<Text bold>some text</Text>))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("with color", () => {
      const tree = renderer
        .create(withTheme(<Text color="success">some text</Text>))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
