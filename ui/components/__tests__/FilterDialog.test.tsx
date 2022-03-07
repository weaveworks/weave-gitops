import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import FilterDialog from "../FilterDialog";

describe("FilterDialog", () => {
  const setActiveFilters = jest.fn();
  const filterList = {
    Name: ["app", "app2", "app3"],
    Status: ["Ready", "Failed"],
    Type: ["Application", "Helm Release"],
  };
  it("should not render when closed", () => {
    render(
      withTheme(
        <FilterDialog
          filterList={filterList}
          onFilterSelect={setActiveFilters}
        />
      )
    );
    expect(screen.queryByText("Name")).toBeNull();
  });
  it("should reveal filter list when open", () => {
    render(
      withTheme(
        <FilterDialog
          open
          filterList={filterList}
          onFilterSelect={setActiveFilters}
        />
      )
    );
    expect(screen.queryByText("Name")).toBeTruthy();
  });
  it("should return a value when a parameter is clicked", () => {
    const onFilterSelect = jest.fn();
    render(
      withTheme(
        <FilterDialog
          open
          filterList={filterList}
          onFilterSelect={onFilterSelect}
        />
      )
    );

    const checkbox1 = document.getElementById("Name.app") as HTMLInputElement;

    expect(checkbox1.checked).toEqual(true);
    fireEvent.click(checkbox1);
    expect(checkbox1.checked).toEqual(false);

    expect(onFilterSelect).toHaveBeenCalledWith({
      Name: ["app2", "app3"],
      Status: ["Ready", "Failed"],
      Type: ["Application", "Helm Release"],
    });

    const checkbox2 = document.getElementById(
      "Type.Application"
    ) as HTMLInputElement;

    expect(checkbox2.checked).toEqual(true);
    fireEvent.click(checkbox2);
    expect(checkbox2.checked).toEqual(false);

    expect(onFilterSelect).toHaveBeenCalledWith({
      Name: ["app2", "app3"],
      Status: ["Ready", "Failed"],
      Type: ["Helm Release"],
    });
  });
});
