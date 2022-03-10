import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import { SortType } from "../DataTable";
import FilterableTable, { filterRows } from "../FilterableTable";
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
});
