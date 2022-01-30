import { fireEvent, render, screen, within } from "@testing-library/react";
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
  describe("pagination", () => {
    it("displays an initial portion of rows based on paginationOptions[0] prop", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
            paginationOptions={[1, 2]}
          />
        )
      );
      expect(screen.getAllByRole("row").length).toEqual(2);
    });
    it("has functioning navigation buttons and display text", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
            paginationOptions={[1, 2]}
          />
        )
      );
      const back = screen.getByLabelText("back one page");
      const skipBack = screen.getByLabelText("skip to first page");
      const forward = screen.getByLabelText("forward one page");
      const skipForward = screen.getByLabelText("skip to last page");
      const displayText = screen.getByText(/1 - 1 out of 3/);
      fireEvent.click(forward);
      expect(displayText.innerHTML).toContain("2 - 2 out of 3");
      fireEvent.click(back);
      expect(displayText.innerHTML).toContain("1 - 1 out of 3");
      fireEvent.click(skipForward);
      expect(displayText.innerHTML).toContain("3 - 3 out of 3");
      fireEvent.click(skipBack);
      expect(displayText.innerHTML).toContain("1 - 1 out of 3");
    });
    it("disables buttons based on page location", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
            paginationOptions={[1, 2]}
          />
        )
      );
      const back = screen.getByLabelText("back one page");
      const skipBack = screen.getByLabelText("skip to first page");
      const forward = screen.getByLabelText("forward one page");
      const skipForward = screen.getByLabelText("skip to last page");
      expect(back.hasAttribute("disabled")).toBeTruthy();
      expect(skipBack.hasAttribute("disabled")).toBeTruthy();
      fireEvent.click(skipForward);
      expect(forward.hasAttribute("disabled")).toBeTruthy();
      expect(skipForward.hasAttribute("disabled")).toBeTruthy();
    });
    it("resets start index to 0 and changes list length on rows per page select change", () => {
      render(
        withTheme(
          <DataTable
            sortFields={["name", "status"]}
            fields={fields}
            rows={rows}
            paginationOptions={[1, 2]}
          />
        )
      );
      const nextPage = screen.getByLabelText("forward one page");
      fireEvent.click(nextPage);
      expect(screen.getAllByRole("row")[1].innerHTML).toMatch(/podinfo/);
      let select;
      screen.getAllByRole("button").forEach((button) => {
        if (button.getAttribute("aria-haspopup")) select = button;
      });
      fireEvent.mouseDown(select);
      const listbox = within(screen.getByRole("listbox"));
      fireEvent.click(listbox.getByText("2"));
      expect(screen.getAllByRole("row")[2].innerHTML).toMatch(/podinfo/);
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
