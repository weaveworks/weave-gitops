import { Meta, Story } from "@storybook/react";
import * as React from "react";
import DataTable, { Props } from "../components/DataTable";

export default {
  title: "DataTable",
  component: DataTable,
  parameters: {
    docs: {
      description: {
        component: "Generates a sorted table based on given fields and rows",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => <DataTable {...args} />;

export const WithData = Template.bind({});
WithData.args = {
  fields: [
    { label: "name", displayLabel: "Name", value: (app) => app.name },
    { label: "type", displayLabel: "Type", value: (app) => app.type },
    {
      label: "overlays",
      displayLabel: "Overlays",
      value: (app) => app.overlays,
    },
    {
      label: "clusters",
      displayLabel: "Clusters",
      value: (app) => app.clusters,
    },
    { label: "status", displayLabel: "Status", value: (app) => app.status },
    { label: "release", displayLabel: "Release", value: (app) => app.release },
  ],
  rows: [
    {
      name: "service",
      status: "Ready on cluster-aaa",
      type: "Helm Release",
      release: "0.1.1",
      overlays: "dev",
      clusters: "1/1",
    },
    {
      name: "auth",
      status: "Failed",
      type: "Application",
      release: "0.2.1",
      overlays: "dev",
      clusters: "1/1",
    },
    {
      name: "app",
      status: "Ready on cluster-aaa",
      type: "Helm Release",
      release: "0.3.1",
      overlays: "prod",
      clusters: "1/1",
    },
    {
      name: "testeroo",
      status: "Ready on cluster-aaa",
      type: "Application",
      release: "1.0",
      overlays: "prod",
      clusters: "1/1",
    },
    {
      name: "another app",
      status: "Ready on cluster-aaa",
      type: "Application",
      release: "1.0",
      overlays: "prod",
      clusters: "1/1",
    },
    {
      name: "z - the last app",
      status: "Ready on cluster-aaa",
      type: "Application",
      release: "1.0",
      overlays: "prod",
      clusters: "1/1",
    },
  ],
  sortFields: ["name", "type", "overlays"],
  paginationOptions: [2, 4],
  className: "",
};
export const NoRows = Template.bind({});
NoRows.args = { ...WithData.args, rows: [] };
