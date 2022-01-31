import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import DataTable from "../DataTable";

describe("DataTable", () => {
  const rows = [
    {
      name: "the-cool-app",
      status: "Ready",
      lastUpdate: "2006-01-02T15:04:05-0700",
    },
    {
      name: "podinfo",
      status: "Failed",
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
      label: "name",
      displayLabel: "Name",
      value: ({ name }) => <a href="/some_url">{name}</a>,
    },
    {
      label: "status",
      displayLabel: "Status",
      value: (v) => v.status,
    },
    {
      label: "lastUpdate",
      displayLabel: "Last Updated",
      value: "lastUpdate",
    },
  ];
  describe("sorting", () => {
    it("initially sorts based on sortFields[0]", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
          />
        )
      );
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/nginx/);
    });
    it("reverses sort on thead click", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
          />
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
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
          />
        )
      );
      const nameButton = screen.getByText("Name");
      fireEvent.click(nameButton);
      const statusButton = screen.getByText("Status");
      fireEvent.click(statusButton);
      const firstRow = screen.getAllByRole("row")[1];
      expect(firstRow.innerHTML).toMatch(/podinfo/);
    });
  });
  describe("snapshots", () => {
    it("renders", () => {
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
