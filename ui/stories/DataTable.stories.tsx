import { Meta, Story } from "@storybook/react";
import * as React from "react";
import DataTable from "../components/DataTable";

export default {
  title: "DataTable",
  component: DataTable,
  argTypes: {
    fields: {
      description:
        "A list of objects with two fields: `label`, which is a string representing the column header, and `value`, which can be a string, or a function that extracts the data needed to fill the table cell",
      defaultValue: [
        { label: "Name", value: (obj) => obj.name },
        { label: "<3s Kubernetes", value: (obj) => `${obj.kub}` },
        { label: "Favorite Food", value: (obj) => obj.food },
      ],
      type: { required: true },
    },
    rows: {
      description:
        "A list of data that will be iterated through to create the columns described in `fields`",
      defaultValue: [
        { name: "Josh", kub: true, food: "Delicious Code" },
        { name: "Evil Josh", kub: false, food: "Sour Patch Kids" },
      ],
      type: { required: true },
    },
    sortFields: {
      description:
        "A list of strings representing the sortable columns of the table, passed into lodash's `_.sortBy`",
      defaultValue: ["name"],
      type: { required: true },
    },
    reverseSort: {
      description: "Indicates whether to reverse the sorted array",
      defaultValue: false,
      type: { required: false },
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
