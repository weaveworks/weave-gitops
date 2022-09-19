import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import MessageBox from "../MessageBox";

describe("MessageBox", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(
          withTheme(
            <MessageBox column align>
              Column and items centered.
            </MessageBox>
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
