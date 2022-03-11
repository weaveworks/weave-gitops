import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import { SortType } from "../DataTable";
import FilterableTable, {
  filterConfigForType,
  filterRows,
} from "../FilterableTable";
import Icon, { IconType } from "../Icon";

describe("FilterableTable", () => {
  const rows = [
    {
      name: "cool",
      type: "foo",
      success: false,
      count: 1,
    },
    {
      name: "slick",
      type: "foo",
      success: true,
      count: 2,
    },
    {
      name: "neat",
      type: "bar",
      success: false,
      count: 500,
    },
    {
      name: "rad",
      type: "baz",
      success: false,
      count: 500,
    },
  ];

  const fields = [
    {
      label: "Name",
      value: "name",
    },
    {
      label: "Type",
      value: "type",
    },
    {
      label: "Status",
      value: "success",
      filterType: SortType.bool,
      displayText: (r) =>
        r.success ? (
          <Icon
            color="success"
            size="base"
            type={IconType.CheckMark}
            text="Successful!"
          />
        ) : (
          <Icon color="alert" size="base" type={IconType.ErrorIcon} />
        ),
    },
    {
      label: "Qty",
      value: "count",
    },
  ];
  describe("filterRows", () => {
    it("shows all rows", () => {
      const filtered = filterRows(rows, {});
      expect(filtered).toHaveLength(4);
    });
    it("filters rows", () => {
      const filtered = filterRows(rows, { name: ["cool"] });
      expect(filtered).toHaveLength(1);
    });
    it("filters rows with more than one value in a filter key", () => {
      const filtered = filterRows(rows, { name: ["cool", "slick"] });
      expect(filtered).toHaveLength(2);
    });
    it("filters rows with multiple filter keys", () => {
      const filtered = filterRows(rows, { name: ["cool"], type: ["bar"] });
      expect(filtered).toHaveLength(2);
      const cool = _.find(filtered, { name: "cool" });
      const neat = _.find(filtered, { name: "neat" });
      const slick = _.find(filtered, { name: "slick" });

      expect(cool).toBeTruthy();
      expect(neat).toBeTruthy();
      expect(slick).toBeFalsy();
    });
    it("filters rows with multiple filters in multiple keys", () => {
      const filtered = filterRows(rows, {
        name: ["cool"],
        type: ["bar", "foo"],
      });
      expect(filtered).toHaveLength(3);

      const filtered2 = filterRows(rows, {
        name: [],
        type: ["baz", "foo"],
      });
      expect(filtered2).toHaveLength(3);

      const neat = _.find(filtered2, { name: "neat" });
      expect(neat).toBeFalsy();
    });
  });

  it("should show all rows", () => {
    render(
      withTheme(<FilterableTable fields={fields} rows={rows} filters={{}} />)
    );

    expect(screen.queryByText("slick")).toBeTruthy();
    expect(screen.queryByText("cool")).toBeTruthy();
  });
  it("should filter rows", () => {
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={{ name: ["slick"] }}
        />
      )
    );

    expect(screen.queryByText("slick")).toBeTruthy();
    expect(screen.queryByText("cool")).toBeFalsy();
  });
  it("should filter on click", () => {
    const initialFilterState = {
      ...filterConfigForType(rows),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
      )
    );

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);

    const tableRows = document.querySelectorAll("tbody tr");

    expect(tableRows).toHaveLength(2);
    expect(tableRows[0].innerHTML).toContain("cool");
    expect(tableRows[1].innerHTML).toContain("slick");

    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();

    const checkbox2 = document.getElementById("type:baz") as HTMLInputElement;
    fireEvent.click(checkbox2);

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(3);
    expect(tableRows2[1].innerHTML).toContain("rad");

    const chip2 = screen.getByText("type:baz");
    expect(chip2).toBeTruthy();
  });
  it("should remove a param when a single chip is clicked", () => {
    const initialFilterState = {
      ...filterConfigForType(rows),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
      )
    );

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);

    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();
    expect(screen.queryByText("Clear All")).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    // TODO: this is probably an a11y problem. The SVG needs "role=button", since it is clickable
    const svgButton = chip1.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    // Should return to all rows being shown
    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(rows.length);

    expect(screen.queryByText("type:foo")).toBeFalsy();
    expect(screen.queryByText("Clear All")).toBeFalsy();
  });
  it("should clear filtering when the `clear all` chip is clicked", () => {
    const initialFilterState = {
      ...filterConfigForType(rows),
    };
    render(
      withTheme(
        <FilterableTable
          fields={fields}
          rows={rows}
          filters={initialFilterState}
          dialogOpen
        />
      )
    );

    const checkbox1 = document.getElementById("type:foo") as HTMLInputElement;
    fireEvent.click(checkbox1);
    const chip1 = screen.getByText("type:foo");
    expect(chip1).toBeTruthy();

    const tableRows1 = document.querySelectorAll("tbody tr");
    expect(tableRows1).toHaveLength(2);

    const checkbox2 = document.getElementById("type:baz") as HTMLInputElement;
    fireEvent.click(checkbox2);
    const chip2 = screen.getByText("type:baz");
    expect(chip2).toBeTruthy();

    const tableRows2 = document.querySelectorAll("tbody tr");
    expect(tableRows2).toHaveLength(3);

    const clearAll = screen.getByText("Clear All");
    // TODO: this is probably an a11y problem. The SVG needs "role=button", since it is clickable
    const svgButton = clearAll.parentElement.getElementsByTagName("svg")[0];
    fireEvent.click(svgButton);

    const tableRows3 = document.querySelectorAll("tbody tr");

    expect(tableRows3).toHaveLength(rows.length);
  });
});
