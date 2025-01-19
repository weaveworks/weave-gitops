import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import Flex from "../Flex";

describe("Flex", () => {
  describe("snapshots", () => {
    it("wide", () => {
      const tree = render(<Flex wide>My Text</Flex>).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("center", () => {
      const tree = render(<Flex center>My Text</Flex>).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("align", () => {
      const tree = render(<Flex align>My Text</Flex>).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("wide center align", () => {
      const tree = render(
        <Flex wide center align>
          My Text
        </Flex>,
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
