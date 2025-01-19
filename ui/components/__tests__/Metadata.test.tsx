import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import Metadata from "../Metadata";

describe("Metadata", () => {
  describe("snapshots", () => {
    it("renders with data", () => {
      const tree = render(
        withTheme(
          <Metadata
            metadata={[
              ["CreatedBy", "Value 4"],
              ["Version", "some version"],
              ["created-by", "Value 2"],
              ["createdBy", "Value 3"],
              ["description", "Value 1"],
              ["html", "<p><b>html</b></p>"],
              ["link-to-google", "https://google.com"],
              ["multi-lines", "This is first line\nThis is second line\n"],
            ]}
            artifactMetadata={[
              ["description", "Value 1"],
              ["html", "<p><b>html</b></p>"],
              ["link-to-google", "https://google.com"],
            ]}
            labels={[
              ["label", "label"],
              ["goose", "goose"],
            ]}
          />,
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("renders nothing without data", () => {
      const tree = render(withTheme(<Metadata />)).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
