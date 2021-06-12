import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import DataTable from "../DataTable";

describe("DataTable", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const rows = [
        {
          name: "my-cool-app",
          status: "Ready",
          lastUpdate: "2006-01-02T15:04:05-0700",
        },
        {
          name: "podinfo",
          status: "Ready",
          lastUpdate: "2006-01-02T15:04:05-0700",
        },
        {
          name: "nginx",
          status: "Ready",
          lastUpdate: "2006-01-02T15:04:05-0700",
        },
      ];

      const fields = [
        {
          label: "Name",
          value: ({ name }) => <a href="/some_url">{name}</a>,
        },
        {
          label: "Status",
          value: (v) => v.status,
        },
        {
          label: "Last Updated",
          value: "lastUpdate",
        },
      ];

      const tree = renderer
        .create(
          withTheme(
            <DataTable sortFields={["name"]} fields={fields} rows={rows} />
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
