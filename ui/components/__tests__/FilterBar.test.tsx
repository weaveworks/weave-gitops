import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import FilterBar from "../FilterBar";

describe("FilterBar", () => {
  const setActiveFilters = jest.fn();
  const filterList = {
    Name: ["app", "app2", "app3"],
    Status: ["Ready", "Failed"],
    Type: ["Application", "Helm Release"],
  };
  it("should initially render button with filter list closed", () => {
    render(
      withTheme(
        <FilterBar
          filterList={filterList}
          activeFilters={[]}
          setActiveFilters={setActiveFilters}
        />
      )
    );
    expect(screen.getByRole("button")).toBeTruthy();
    expect(screen.queryByText("Name")).toBeNull();
  });
  it("should reveal filter list on icon click", () => {
    render(
      withTheme(
        <FilterBar
          filterList={filterList}
          activeFilters={[]}
          setActiveFilters={setActiveFilters}
        />
      )
    );
    const icon = screen.getAllByRole("button")[0];
    fireEvent.click(icon);
    expect(screen.queryByText("Name")).toBeTruthy();
  });
});
