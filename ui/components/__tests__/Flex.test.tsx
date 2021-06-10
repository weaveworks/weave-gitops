import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import Flex from "../Flex";

describe("Flex", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(
          <Flex wide center align>
            Aligned and Centered!
          </Flex>
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
