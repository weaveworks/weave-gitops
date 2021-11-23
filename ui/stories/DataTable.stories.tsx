import { Meta, Story } from "@storybook/react";
import * as React from "react";
import DataTable from "../components/DataTable";

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
  argTypes: {
    fields: {
      description:
        "A list of objects with two fields: `label`, which is a string representing the column header, and `value`, which can be a string, or a function that extracts the data needed to fill the table cell",
      defaultValue: [
        { label: "Name", value: (app) => app.name },
        { label: "Status", value: (app) => app.status },
        { label: "Type", value: (app) => app.type },
        { label: "Version", value: (app) => app.version },
        { label: "Last Commit", value: (app) => app.last_commit },
        { label: "Namespace", value: (app) => app.namespace },
      ],
      type: { required: true, summary: "array" },
    },
    rows: {
      description:
        "A list of data that will be iterated through to create the columns described in `fields`",
      defaultValue: [
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
      type: { required: true, summary: "array" },
    },
    sortFields: {
      description:
        "A list of strings representing the sortable columns of the table, passed into lodash's `_.sortBy`",
      defaultValue: ["name"],
      type: { required: true, summary: "array" },
    },
    reverseSort: {
      description: "Indicates whether to reverse the sorted array",
      defaultValue: false,
      type: { required: false, name: "boolean", summary: "boolean" },
      control: "boolean",
    },
    className: {
      description: "CSS MUI Overrides or other styling",
      type: { summary: "string", name: "string", required: false },
    },
  },
} as Meta;

const Template: Story = (args) => <DataTable {...args} />;

export const Default = Template.bind({});
export const NoRows = Template.bind({});
NoRows.args = { rows: [] };
