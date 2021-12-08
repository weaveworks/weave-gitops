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
    { label: "Name", value: (app) => app.name },
    { label: "Status", value: (app) => app.status },
    { label: "Type", value: (app) => app.type },
    { label: "Version", value: (app) => app.version },
    { label: "Last Commit", value: (app) => app.last_commit },
    { label: "Namespace", value: (app) => app.namespace },
  ],
  rows: [
    {
      name: "Real App",
      status: "Ready on cluster-aaa",
      type: "Application",
      version: "0.1.1",
      last_commit: "@joshri commit 000000",
      namespace: "default",
    },
    {
      name: "Other Real App",
      status: "Ready on cluster-aaa",
      type: "Application",
      version: "0.2.1",
      last_commit: "@joshri commit 000000",
      namespace: "default",
    },
    {
      name: "This Is My Third Real App",
      status: "Ready on cluster-aaa",
      type: "Application",
      version: "0.3.1",
      last_commit: "@joshri commit 000000",
      namespace: "default",
    },
  ],
  sortFields: ["name"],
  reverseSort: false,
  className: "",
};
export const NoRows = Template.bind({});
NoRows.args = { ...WithData.args, rows: [] };
