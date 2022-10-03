import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import Flex from "../Flex";

describe("Flex", () => {
  describe("snapshots", () => {
    it("wide", () => {
      const tree = renderer.create(<Flex wide>My Text</Flex>).toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("center", () => {
      const tree = renderer.create(<Flex center>My Text</Flex>).toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("align", () => {
      const tree = renderer.create(<Flex align>My Text</Flex>).toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("wide center align", () => {
      const tree = renderer
        .create(
          <Flex wide center align>
            My Text
          </Flex>
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
