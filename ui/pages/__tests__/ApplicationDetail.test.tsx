import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import "jest-styled-components";
import * as React from "react";
import { act } from "react-dom/test-utils";
import { AppProps } from "../../contexts/AppContext";
import {
  GitProvider,
  ListCommitsResponse,
  ParseRepoURLResponse,
  SyncApplicationResponse,
} from "../../lib/api/applications/applications.pb";
import { createMockClient, withContext, withTheme } from "../../lib/test-utils";
import ApplicationDetail from "../ApplicationDetail";

describe("ApplicationDetail", () => {
  describe("Sync App Button", () => {
    const apiOverrides = {
      GetApplication: () => ({
        application: {
          name: "pod-info",
          namespace: "wego-systems",
        },
      }),
      ListCommits: (): ListCommitsResponse => ({
        commits: [
          {
            hash: "123abc",
            author: "Example User",
            date: "2021-09-10T23:45:09Z",
          },
        ],
      }),
      ParseRepoURL: (): ParseRepoURLResponse => ({
        provider: GitProvider.GitHub,
        owner: "someone",
      }),
      SyncApplication: (): SyncApplicationResponse => {
        return {
          success: true,
        };
      },
    };

    it("should exist on page", async () => {
      render(
        withTheme(
          withContext(
            <ApplicationDetail name="pod-info" />,
            "/application_detail",
            { applicationsClient: createMockClient(apiOverrides) }
          )
        )
      );
      expect(await screen.findByText("Sync App")).toBeTruthy();
    });
    it("should call a sync request", async () => {
      const promise = Promise.resolve();
      const syncMock = jest.fn();
      const mockClient = createMockClient(apiOverrides);

      mockClient.SyncApplication = (req) => {
        syncMock(req);
        return new Promise((accept) => {
          accept({ success: true });
        });
      };

      await act(async () => {
        render(
          withTheme(
            withContext(
              <ApplicationDetail name="pod-info" />,
              "/application_detail",
              { applicationsClient: mockClient }
            )
          )
        );
      });

      const button = await (
        await screen.findByText("Sync App")
      ).closest("button");

      fireEvent(button, new MouseEvent("click", { bubbles: true }));

      expect(syncMock).toHaveBeenCalledWith({
        name: "pod-info",
        namespace: "wego-systems",
      });
      await waitFor(() => promise);
    });
    it("should notify user on success", async () => {
      const promise = Promise.resolve();
      const props: AppProps = {
        applicationsClient: createMockClient(apiOverrides),
        notifySuccess: jest.fn(),
      };

      await act(async () => {
        render(
          withTheme(
            withContext(
              <ApplicationDetail name="pod-info" />,
              "/application_detail",
              props
            )
          )
        );
      });

      const button = await (
        await screen.findByText("Sync App")
      ).closest("button");

      fireEvent(button, new MouseEvent("click", { bubbles: true }));
      await waitFor(() => promise);
      expect(props.notifySuccess).toHaveBeenCalledTimes(1);
    });
    it("should notify user on failure", async () => {
      const errorText = "uh-oh";
      const mockClient = createMockClient(apiOverrides);
      mockClient.SyncApplication = () =>
        new Promise((_, reject) => reject({ message: errorText }));

      await act(async () => {
        render(
          withTheme(
            withContext(
              <ApplicationDetail name="pod-info" />,
              "/application_detail",
              { applicationsClient: mockClient }
            )
          )
        );
      });

      const button = await (
        await screen.findByText("Sync App")
      ).closest("button");
      fireEvent(button, new MouseEvent("click", { bubbles: true }));
      await screen.findByText(errorText);
      await waitFor(() => mockClient.SyncApplication);
    });
  });
});
