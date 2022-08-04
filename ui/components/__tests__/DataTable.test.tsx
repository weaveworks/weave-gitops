import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import DataTable, { SortType } from "../DataTable";

describe("DataTable", () => {
  const rows = [
    {
      name: "the-cool-app",
      status: true,
      lastUpdate: "2005-01-02T15:04:05-0700",
      lastSyncedAt: 1000,
    },
    {
      name: "podinfo",
      status: false,
      lastUpdate: "2006-01-02T15:04:05-0700",
      lastSyncedAt: 2000,
    },
    {
      name: "nginx",
      status: "Ready",
      lastUpdate: "2004-01-02T15:04:05-0700",
      lastSyncedAt: 3000,
      suspended: true,
    },
  ];

  const fields = [
    {
      label: "Name",
      value: ({ name }) => <a href="/some_url">{name}</a>,
      sortType: SortType.string,
      sortValue: ({ name }) => name,
    },
    {
      label: "Status",
      value: "status",
      sortType: SortType.number,
      sortValue: ({ status, suspended }) => {
        if (suspended) return 2;
        if (status) return 3;
        else return 1;
      },
    },
    {
      label: "Last Updated",
      value: "lastUpdate",
      sortType: SortType.date,
      sortValue: ({ lastUpdate }) => lastUpdate,
    },
    {
      label: "Last Synced At",
      value: "lastSyncedAt",
      sortType: SortType.number,
      sortValue: ({ lastSyncedAt }) => lastSyncedAt,
    },
  ];

  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(withTheme(<DataTable fields={fields} rows={rows} />))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
