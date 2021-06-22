import { render, screen } from "@testing-library/react";
import _ from "lodash";
import * as React from "react";
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

  it("lists applications", async () => {
    const name = "some app";
    const mockResponses = {
      ListApplications: { applications: [{ name }] },
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

    expect((await screen.findByTestId(name)).textContent).toEqual(name);
  });
});
