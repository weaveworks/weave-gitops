import { render, screen } from "@testing-library/react";
import _ from "lodash";
import * as React from "react";
import { Applications } from "../../lib/api/applications/applications.pb";
import { withContext } from "../../lib/test-utils";
import useApplications from "../applications";

describe("useApplications", () => {
  let container;
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });
  afterEach(() => {
    document.body.removeChild(container);
    container = null;
  });

  it("lists applications", () => {
    const mockResponses: typeof Applications = {
      ListApplications: () => [],
    };
    const TestComponent = () => {
      const { applications } = useApplications();

      return (
        <ul>
          {_.map(applications, (a) => (
            <li key={a.name} data-testid={a.name}>
              {a.name}
            </li>
          ))}
        </ul>
      );
    };

    render(withContext(TestComponent, `/`, mockResponses));

    expect(screen.getByTestId(id).textContent).toEqual(id);
  });
});
