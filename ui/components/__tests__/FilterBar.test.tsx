import { render, screen } from "@testing-library/react";
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
  it("should initially render clear all chip with filter list closed", () => {
    render(
      withTheme(
        <FilterBar
          filterList={filterList}
          activeFilters={[]}
          setActiveFilters={setActiveFilters}
        />
      )
    );
    expect(screen.getByText("Clear All")).toBeTruthy();
    expect(screen.queryByText("Name")).toBeNull();
  });
  it("should reveal/close filter list on icon click", () => {
    ("");
  });
  it("should should add filter to chips on checkbox click", () => {
    ("");
  });
  it("should clear all filters on click on clear all chip", () => {
    ("");
  });
  it("should change filter list based on search input", () => {
    ("");
  });
});
