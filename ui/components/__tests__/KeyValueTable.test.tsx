import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import KeyValueTable from "../KeyValueTable";

describe("KeyValueTable", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(
          withTheme(
            <KeyValueTable
              columns={4}
              pairs={[
                { key: "name", value: "my thing" },
                {
                  key: "status",
                  value: "ok",
                },
                {
                  key: "Last Updated",
                  value: "2006-01-02T15:04:05-0700",
                },
              ]}
            />
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
