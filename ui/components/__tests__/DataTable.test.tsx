import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withContext, withTheme } from "../../lib/test-utils";
import DataTable from "../DataTable";

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
      sortValue: ({ name }) => name,
      defaultSort: true,
    },
    {
      label: "Status",
      value: "status",
      sortValue: ({ status, suspended }) => {
        if (suspended) return 2;
        if (status) return 3;
        else return 1;
      },
    },
    {
      label: "Last Updated",
      value: "lastUpdate",
      sortValue: ({ lastUpdate }) => lastUpdate,
    },
    {
      label: "Last Synced At",
      value: "lastSyncedAt",
      sortValue: ({ lastSyncedAt }) => lastSyncedAt,
    },
  ];

  describe("sorting", () => {
    it("initially sorts based on defaultSort", () => {
      render(
        withTheme(
          withContext(
            <DataTable fields={fields} rows={rows} />,
            "/applications",
            {}
          )
        )
      );
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/nginx/);
    });
    it("reverses sort on thead click", () => {
      render(
        withTheme(
          withContext(
            <DataTable fields={fields} rows={rows} />,
            "/applications",
            {}
          )
        )
      );

      const nameButton = screen.getByText("Name");
      fireEvent.click(nameButton);
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/the-cool-app/);
    });
    it("resets reverseSort and switches sort column on different thead click", () => {
      render(
        withTheme(
          withContext(
            <DataTable fields={fields} rows={rows} />,
            "/applications",
            {}
          )
        )
      );
      const nameButton = screen.getByText("Name");
      fireEvent.click(nameButton);
      const statusButton = screen.getByText("Status");
      fireEvent.click(statusButton);
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/podinfo/);
    });
    it("should render text when rows are empty", () => {
      render(
        withTheme(
          withContext(
            <DataTable fields={fields} rows={[]} />,
            "/applications",
            {}
          )
        )
      );
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/No/);
    });
    it("sorts by value when no sortValue property exists", () => {
      const rows = [
        {
          name: "b",
        },
        {
          name: "c",
        },
        {
          name: "a",
        },
      ];

      const fields = [
        {
          label: "Name",
          value: "name",
        },
      ];

      render(
        withTheme(
          withContext(
            <DataTable fields={fields} rows={rows} />,
            "/applications",
            {}
          )
        )
      );

      let firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/a/);

      const nameButton = screen.getByText("Name");
      fireEvent.click(nameButton);

      firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/c/);
    });
  });

  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <DataTable fields={fields} rows={rows} />,
              "/applications",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
