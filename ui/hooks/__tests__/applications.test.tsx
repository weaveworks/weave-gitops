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
      GetUser: { user: { email: "user@example.com" } },
    };
    const TestComponent = () => {
      const { listApplications, applications } = useApplications();

      React.useEffect(() => {
        listApplications();
      }, []);

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
  it("get application", async () => {
    const url = "example.com/somepath";
    const mockResponses = {
      ListApplications: { applications: [{ name: "some-name" }] },
      GetApplication: { application: { url } },
      GetUser: { user: { email: "user@example.com" } },
    };
    const TestComponent = () => {
      const { currentApplication: app, getApplication } = useApplications();

      React.useEffect(() => {
        getApplication("my-app");
      }, []);

      return <p data-testid="url">{app.url}</p>;
    };

    render(withContext(TestComponent, `/`, mockResponses));

    expect((await screen.findByTestId("url")).textContent).toEqual(url);
  });
});
