import { render, screen } from "@testing-library/react";
import _ from "lodash";
import * as React from "react";
import { Application } from "../../lib/api/applications/applications.pb";
import { createMockClient, withContext } from "../../lib/test-utils";
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
      ListApplications: () => ({
        applications: [{ name }],
      }),
    };
    const TestComponent = () => {
      const { listApplications } = useApplications();
      const [applications, setApplications] = React.useState<Application[]>([]);

      React.useEffect(() => {
        listApplications().then((res) => setApplications(res as Application[]));
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

    render(
      withContext(TestComponent, `/`, {
        applicationsClient: createMockClient(mockResponses),
      })
    );

    expect((await screen.findByTestId(name)).textContent).toEqual(name);
  });
  it("get application", async () => {
    const url = "example.com/somepath";
    const mockResponses = {
      ListApplications: () => ({ applications: [{ name: "some-name" }] }),
      GetApplication: () => ({ application: { url } }),
    };
    const TestComponent = () => {
      const { getApplication } = useApplications();
      const [app, setApp] = React.useState({} as any);

      React.useEffect(() => {
        getApplication("my-app").then((a) => setApp(a as any));
      }, []);

      return <p data-testid="url">{app.url}</p>;
    };

    render(
      withContext(TestComponent, `/`, {
        applicationsClient: createMockClient(mockResponses),
      })
    );

    expect((await screen.findByTestId("url")).textContent).toEqual(url);
  });
});
