import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import Text from "../Text";

describe("Text", () => {
  describe("snapshots", () => {
    it("normal", () => {
      const tree = render(withTheme(<Text>some text</Text>)).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("bold", () => {
      const tree = render(withTheme(<Text bold>some text</Text>)).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("with color", () => {
      const tree = render(
        withTheme(<Text color="successOriginal">some text</Text>),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
