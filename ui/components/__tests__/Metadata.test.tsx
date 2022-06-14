import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import Metadata from "../Metadata";

describe("snapshots", () => {
  it("renders with data", () => {
    const tree = renderer
      .create(
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
          />
        )
      )
      .toJSON();
    expect(tree).toMatchSnapshot();
  });
  it("renders nothing without data", () => {
    const tree = renderer.create(withTheme(<Metadata />)).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
