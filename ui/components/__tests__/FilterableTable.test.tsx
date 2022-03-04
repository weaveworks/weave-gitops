import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import { SortType } from "../DataTable";
import FilterableTable from "../FilterableTable";
import Icon, { IconType } from "../Icon";

describe("FilterableTable", () => {
  const rows = [
    {
      name: "cool",
      success: false,
      count: 1,
    },
    {
      name: "slick",
      success: true,
      count: 2,
    },
    {
      name: "neat",
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
